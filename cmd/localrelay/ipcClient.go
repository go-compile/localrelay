package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

// commandDataIPC will send a command with a data section
func commandDataIPC(w io.Writer, id uint8, data []byte) error {
	// calculate packet length
	payloadLen := make([]byte, 2)
	binary.BigEndian.PutUint16(payloadLen, uint16(len(data)+3))

	if _, err := w.Write(payloadLen); err != nil {
		return err
	}

	if _, err := w.Write([]byte{id}); err != nil {
		return err
	}

	// write buf len
	lenBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(lenBuf, uint16(len(data)))

	if _, err := w.Write(lenBuf); err != nil {
		return err
	}

	if _, err := w.Write([]byte(data)); err != nil {
		return err
	}

	return nil
}

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

			errlenBuf := make([]byte, 2)
			if _, err := conn.Read(errlenBuf); err != nil {
				return err
			}

			fmt.Println("1")
			msg := make([]byte, binary.BigEndian.Uint16(errlenBuf))
			if _, err := conn.Read(msg); err != nil {
				return err
			}

			fmt.Println(string(msg))
		case 1:
			fmt.Printf("[Info] Relay %q has been started.\n", relay)
		case 2:
			fmt.Printf("[Info] Relay %q is already running.\n", relay)
		case 3:
			fmt.Printf("[Info] Relay %q does not exist.\n", relay)
		case 4:
			fmt.Printf("[Info] Relay %q errored when creating.\n", relay)
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

func stopRelay(relayName string) error {
	conn, err := IPCConnect()
	if err != nil {
		return err
	}

	defer conn.Close()

	if err := commandDataIPC(conn, daemonStop, []byte(relayName)); err != nil {
		return nil
	}

	response := make([]byte, 1)
	if _, err := conn.Read(response); err != nil {
		return err
	}

	switch response[0] {
	case 0:
		fmt.Println("Failed to stop relay.")
	case 1:
		fmt.Printf("Relay %q has been stopped.\n", relayName)
	case 3:
		fmt.Printf("Relay %q is not running.\n", relayName)
	case 4:
		fmt.Printf("Remote daemon failed to parse command.\n")
	default:
		fmt.Printf("Unknown response %d.\n", response[0])
	}

	return nil
}
