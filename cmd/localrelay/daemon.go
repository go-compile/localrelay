package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"

	"github.com/go-compile/localrelay"
)

const (
	serviceName = "localrelay-pipe"
)

const (
	// daemonStatus is used to request the relay state
	daemonStatus uint8 = iota
	// daemonStop will close all the relays and stop the daemon
	daemonStop
)

type status struct {
	Relays []localrelay.Relay
	Pid    int
}

func readCommand(conn net.Conn) ([]byte, error) {
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

func parseCommand(conn net.Conn, payload []byte) (uint8, error) {
	r := bytes.NewBuffer(payload)

	commandID, err := r.ReadByte()
	if err != nil {
		return 0, err
	}

	return commandID, nil
}

func fork() error {
	s, err := getDaemonStatus()
	if err == nil {
		fmt.Printf("[Fatal] Localrelay already running on PID: %d\n", s.Pid)
		for num, r := range s.Relays {
			fmt.Printf(" %.2d: %s  %s -> %s\n", num+1, r.Name, r.Host, r.ForwardAddr)
		}

		return nil
	}

	fmt.Println("Daemon not running, forking & starting relay daemon now")

	cmd := exec.Command(os.Args[0], append([]string{"-" + forkIdentifier}, os.Args[1:]...)...)
	err = cmd.Start()
	if err != nil {
		return err
	}

	fmt.Printf("[Info] Relays running in background on PID: %d\n", cmd.Process.Pid)
	return nil
}
