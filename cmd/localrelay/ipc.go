package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"github.com/go-compile/localrelay"
	"github.com/naoina/toml"
	"github.com/pkg/errors"
)

type daemon struct{}

const (
	serviceName        = "localrelayd"
	ipcSocket          = "localrelay.ipc.socket"
	serviceDescription = "Localrelay daemon relay runner"
)

const (
	daemonRun uint8 = iota
	daemonStatus
	daemonStop
	daemonConns
	daemonDropAll   //TODO
	daemonDropRelay //TODO
	daemonDropIP    //TODO
	daemonDropAddr  //TODO

	maxErrors = 40
)

var (
	ipcListener io.Closer
)

func readCommand(conn io.Reader) ([]byte, error) {
	buf := make([]byte, 2)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint16(buf)
	payload := make([]byte, length)

	if _, err := io.ReadFull(conn, payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func parseCommand(payload []byte) (uint8, []byte, error) {
	r := bytes.NewBuffer(payload)

	commandID, err := r.ReadByte()
	if err != nil {
		return 0, nil, err
	}

	// uint16
	dataLenBuf := make([]byte, 2)
	if _, err := io.ReadFull(r, dataLenBuf); err != nil {
		return 0, nil, err
	}

	dataLen := binary.BigEndian.Uint16(dataLenBuf)
	if dataLen == 0 {
		return commandID, nil, nil
	}

	data := make([]byte, dataLen)
	if _, err := io.ReadFull(r, data); err != nil {
		return 0, nil, err
	}

	return commandID, data, nil
}

// handleConn takes a conn and handles each command
func handleConn(conn io.ReadWriteCloser, l io.Closer) {
	defer conn.Close()

	// if client causes more than x errors drop conn
	for i := 0; i < maxErrors; {
		// run in loop and handle errors here
		if err := ipcLoop(conn); err != nil {
			i++ //increase error counter

			// connection has closed
			if err == io.EOF {
				return
			}

			log.Printf("[Error] IPC: %s\n", err)

			// if connection closed quit
			if err == net.ErrClosed || err == io.EOF {
				break
			}
		}
	}
}

func ipcLoop(conn io.ReadWriteCloser) error {
	// read command payload
	cmdBuf, err := readCommand(conn)
	if err != nil {
		return err
	}

	cmdID, data, err := parseCommand(cmdBuf)
	if err != nil {
		conn.Write([]byte{4})
		return err
	}

	switch cmdID {
	case daemonStop:
		relayName := string(data)

		var relay *localrelay.Relay
		for _, r := range runningRelays() {
			if r.Name == strings.ToLower(relayName) {
				relay = r
				break
			}
		}

		// relay not found
		if relay == nil {
			// send not found response
			conn.Write([]byte{3})
			return nil
		}

		if err := relay.Close(); err != nil {
			conn.Write([]byte{0})
			return err
		}

		// send success
		conn.Write([]byte{1})
	case daemonRun:
		relayFile := string(data)
		exists, err := pathExists(relayFile)
		if err != nil {
			conn.Write([]byte{4})
			return err
		}

		if !exists {
			conn.Write([]byte{3})
			return os.ErrNotExist
		}

		f, err := os.Open(relayFile)
		if err != nil {
			return errors.Wrapf(err, "file:%q", relayFile)
		}

		var relay Relay
		if err := toml.NewDecoder(f).Decode(&relay); err != nil {
			f.Close()
			return err
		}

		f.Close()

		if isRunning(relay.Name) {
			conn.Write([]byte{2})
			return nil
		}

		if err := launchRelays([]Relay{relay}, false); err != nil {
			msgLen := make([]byte, 2)
			binary.BigEndian.PutUint16(msgLen, uint16(len(err.Error())))

			conn.Write([]byte{0, msgLen[0], msgLen[1]})
			conn.Write([]byte(err.Error()))

			return err
		}

		// send success response
		conn.Write([]byte{1})
	case daemonStatus:
		respBuf := bytes.NewBuffer(nil)

		relayMetrics := make(map[string]metrics)

		relays := runningRelaysCopy()
		for _, r := range relays {
			active, total := r.Metrics.Connections()
			relayMetrics[r.Name] = metrics{
				In:            r.Metrics.Download(),
				Out:           r.Metrics.Upload(),
				Active:        active,
				DialAvg:       r.DialerAvg(),
				TotalConns:    total,
				TotalRequests: r.Metrics.Requests(),
			}
		}

		json.NewEncoder(respBuf).Encode(&status{
			Relays:  relays,
			Pid:     os.Getpid(),
			Version: VERSION,
			Started: daemonStarted.Unix(),

			Metrics: relayMetrics,
		})

		lenbuf := make([]byte, 2)
		binary.BigEndian.PutUint16(lenbuf, uint16(respBuf.Len()))

		conn.Write(lenbuf)
		conn.Write(respBuf.Bytes())
	case daemonConns:
		respBuf := bytes.NewBuffer(nil)

		relayConns := make([]connection, 0, 200)

		relays := runningRelaysCopy()
		for _, r := range relays {
			for _, conn := range r.GetConns() {

				relayConns = append(relayConns, connection{
					LocalAddr:  conn.Conn.LocalAddr().String(),
					RemoteAddr: conn.Conn.RemoteAddr().String(),
					Network:    conn.Conn.LocalAddr().Network(),

					RelayName:     r.Name,
					RelayHost:     r.Host,
					ForwardedAddr: conn.RemoteAddr,

					Opened: conn.Opened.Unix(),
				})
			}
		}

		json.NewEncoder(respBuf).Encode(relayConns)

		lenbuf := make([]byte, 2)
		binary.BigEndian.PutUint16(lenbuf, uint16(respBuf.Len()))

		conn.Write(lenbuf)
		conn.Write(respBuf.Bytes())

	case daemonDropAll:
		relays := runningRelaysCopy()
		// iterate through all relays and close every connection
		for _, r := range relays {
			for _, conn := range r.GetConns() {
				go conn.Conn.Close()
			}
		}

		// return success
		conn.Write([]byte{0, 0})
	default:
		// send unsuccessful response
		msg := "Unknown command"
		msgLen := make([]byte, 2)
		binary.BigEndian.PutUint16(msgLen, uint16(len(msg)))

		conn.Write([]byte{0, msgLen[0], msgLen[1]})
		conn.Write([]byte(msg))
	}

	return nil
}
