package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"

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
	client, conn, err := IPCConnect()
	if err != nil {
		return err
	}

	defer conn.Close()

	for _, relay := range relays {
		// make post request to run relay. Use strconv instead of json encoding for performance
		resp, err := client.Post("http://lr/run", "application/json", bytes.NewBuffer([]byte("["+strconv.Quote(relay)+"]")))
		if err != nil {
			return err
		}

		var response msgResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return err
		}

		fmt.Println(response.Message)
	}

	return nil
}

func serviceStatus() (*status, error) {
	client, conn, err := IPCConnect()
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	resp, err := client.Get("http://lr/status")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("failed to fetch status")
	}

	var status status
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, err
	}

	return &status, nil
}

func stopRelay(relayName string) error {
	client, conn, err := IPCConnect()
	if err != nil {
		return err
	}

	defer conn.Close()

	resp, err := client.Get("http://lr/stop/" + url.PathEscape(relayName))
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case 200:
		fmt.Printf("Relay %q has been stopped.\n", relayName)
	case 500:
		fmt.Println("Failed to stop relay.")
	case 404:
		fmt.Printf("Relay not found.\n")
	default:
		fmt.Printf("Unknown response %d.\n", resp.StatusCode)
	}

	return nil
}

func activeConnections() ([]connection, error) {
	client, conn, err := IPCConnect()
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	resp, err := client.Get("http://lr/connections")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("failed to fetch connections")
	}

	var pool []connection
	if err := json.NewDecoder(resp.Body).Decode(&pool); err != nil {
		return nil, err
	}

	return pool, nil
}

func dropAll() error {
	// conn, err := IPCConnect()
	// if err != nil {
	// 	return errors.Wrap(err, "connecting to IPC")
	// }

	// defer conn.Close()

	// _, err = conn.Write([]byte{0, 3, daemonDropAll, 0, 0})
	// if err != nil {
	// 	return err
	// }

	// _, err = readCommand(conn)
	// if err != nil {
	// 	return errors.Wrap(err, "reading from ipc conn")
	// }

	return nil
}
