package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"time"

	"github.com/go-compile/localrelay"
	"github.com/pkg/errors"
)

const (
	serviceName = "com.go-compile.localrelay.ipc.clipipe"
)

const (
	// daemonStatus is used to request the relay state
	daemonStatus uint8 = iota
	// daemonStop will close all the relays and stop the daemon
	daemonStop

	daemonFork
	daemonMetrics
)

var (
	// ErrIPCShutdownFail is returned when the daemon fails to shutdown when being
	// requested via IPC
	ErrIPCShutdownFail = errors.New("failed to shutdown daemon process via IPC")

	// ErrIPCForkFail is returned when trying to re-fork the daemon process
	ErrIPCForkFail = errors.New("ipc fork failed")

	ipcTimeout = time.Second

	// daemonStarted stores the time when the daemon was created
	daemonStarted time.Time
)

type status struct {
	Relays  []localrelay.Relay
	Pid     int
	Version string
	// Metrics contains relay name as the index
	Metrics map[string]metrics
	// Started is a unix timestamp of when the daemon was created
	Started int64
}

type metrics struct {
	In, Out, Active, DialAvg int
	TotalConns               uint64
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

// runFork is used when the user executed the run command
func runFork() error {
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

func fork() error {
	cmd := exec.Command(os.Args[0], append([]string{"-" + forkIdentifier}, os.Args[1:]...)...)
	err := cmd.Start()
	if err != nil {
		return err
	}

	return nil
}
