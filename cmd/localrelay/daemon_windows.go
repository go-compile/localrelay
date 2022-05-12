package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"gopkg.in/natefinch/npipe.v2"
)

func getDaemonStatus() (*status, error) {
	fmt.Println("Attempting to connect to daemon")
	conn, err := npipe.DialTimeout(`\\.\pipe\`+serviceName, time.Second*2)
	if err != nil {
		return nil, err
	}
	fmt.Println("Connected to daemon")

	_, err = conn.Write([]byte{0, 1, daemonStatus})
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

func launchDaemon() {
	conn, err := npipe.DialTimeout(`\\.\pipe\`+serviceName, time.Second*2)
	if err == nil {
		conn.Close()
		fmt.Println("Localrelay service already running.")

		os.Exit(0)
	}

	go startDaemon()
}

func startDaemon() error {
	l, err := npipe.Listen(`\\.\pipe\` + serviceName)
	if err != nil {
		return err
	}

	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			if err == net.ErrClosed {
				return err
			}

			continue
		}

		go handleDaemonConn(conn, l)
	}
}

func handleDaemonConn(conn net.Conn, l *npipe.PipeListener) {
	defer conn.Close()

	for {
		payload, err := readCommand(conn)
		if err != nil {
			return
		}

		cmdID, err := parseCommand(conn, payload)
		if err != nil {
			return
		}

		switch cmdID {
		case daemonStatus:
			respBuf := bytes.NewBuffer(nil)

			json.NewEncoder(respBuf).Encode(&status{
				Relays: runningRelaysCopy(),
				Pid:    os.Getpid(),
			})

			lenbuf := make([]byte, 2)
			binary.BigEndian.PutUint16(lenbuf, uint16(respBuf.Len()))

			conn.Write(lenbuf)
			conn.Write(respBuf.Bytes())

		case daemonStop:
			l.Close()

			for _, r := range runningRelays() {
				log.Printf("[Info] Closing relay: %s\n", r.Name)
				if err := r.Close(); err != nil {
					log.Printf("[Error] Closing relay: %s with error: %s\n", r.Name, err)
				}
			}

			log.Printf("[Info] All relays closed:\n")
			os.Exit(0)
		}
	}
}
