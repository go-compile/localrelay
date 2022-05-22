package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"net"
	"os"

	"github.com/naoina/toml"
	"github.com/pkg/errors"
)

type daemon struct{}

const (
	serviceName        = "Localrelay Service"
	ipcSocket          = "com.go-compile.localrelay.ipc.clipipe"
	serviceDescription = "Localrelay daemon relay runner"
)

const (
	daemonRun uint8 = iota

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
		return err
	}

	switch cmdID {
	case daemonRun:
		relayFile := string(data)
		exists, err := pathExists(relayFile)
		if err != nil {
			return err
		}

		if !exists {
			conn.Write([]byte{0})
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

		if err := launchRelays([]Relay{relay}, false); err != nil {
			return err
		}

		// send success response
		conn.Write([]byte{1})
	default:
		// send unsuccessful response
		conn.Write([]byte{0})
	}

	return nil
}
