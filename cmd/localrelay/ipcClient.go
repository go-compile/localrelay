package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

// serviceRun takes paths to relay config files and then connects via IPC to
// instruct the service to run these relays
func serviceRun(relays []string) error {
	conn, err := IPCConnect()
	if err != nil {
		return errors.Wrap(err, "connecting to IPC")
	}

	defer conn.Close()

	for _, relay := range relays {
		buf := bytes.NewBuffer(nil)
		buf.Write([]byte{daemonRun})

		lenBuf := make([]byte, 2)
		binary.BigEndian.PutUint16(lenBuf, uint16(len(relay)))

		buf.Write(lenBuf)
		buf.Write([]byte(relay))

		payloadLen := make([]byte, 2)
		binary.BigEndian.PutUint16(payloadLen, uint16(buf.Len()))

		conn.Write(payloadLen)
		conn.Write(buf.Bytes())

		response := make([]byte, 1)
		_, err := conn.Read(response)
		if err != nil {
			return errors.Wrap(err, "reading from ipc conn")
		}

		switch response[0] {
		case 0:
			fmt.Printf("[Error] Relay %q could not be started.\n", relay)
		case 1:
			fmt.Printf("[Info] Relay %q has been started.\n", relay)
		case 2:
			fmt.Printf("[Info] Relay %q is already running.\n", relay)
		}
	}

	return nil
}

func serviceStatus() (*status, error) {
	conn, err := IPCConnect()
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	_, err = conn.Write([]byte{0, 3, daemonStatus, 0, 0})
	if err != nil {
		return nil, err
	}

	payload, err := readCommand(conn)
	if err != nil {
		return nil, err
	}

	var s status
	if err := json.Unmarshal(payload, &s); err != nil {
		return nil, err
	}

	return &s, nil
}
