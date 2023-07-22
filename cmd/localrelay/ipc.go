package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"

	"github.com/valyala/fasthttp"
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
	daemonDropAll
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
func handleConn(conn net.Conn, srv *fasthttp.Server, l io.Closer) {
	defer conn.Close()

	srv.ServeConn(conn)
}

func ipcLoop(conn io.ReadWriteCloser) error {
	// read command payload
	cmdBuf, err := readCommand(conn)
	if err != nil {
		return err
	}

	cmdID, _, err := parseCommand(cmdBuf)
	if err != nil {
		conn.Write([]byte{4})
		return err
	}

	switch cmdID {
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
