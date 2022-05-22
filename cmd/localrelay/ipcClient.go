package main

import (
	"bytes"
	"encoding/binary"
	"log"
)

// serviceRun takes paths to relay config files and then connects via IPC to
// instruct the service to run these relays
func serviceRun(relays []string) error {
	conn, err := IPCConnect()
	if err != nil {
		return err
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
			return err
		}

		// TODO: add "already running" error code
		if response[0] != 1 {
			log.Printf("[Error] Relay %q could not be started\n", relay)
		} else {
			log.Printf("[Info] Relay %q has been started\n", relay)
		}
	}

	return nil
}
