package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
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

	p, err := fork()
	if err != nil {
		return err
	}

	fmt.Printf("[Info] Relays running in background on PID: %d\n", p.Pid)
	return nil
}

func fork() (*os.Process, error) {

	// srv, err := daemon.New("localrelay", "localrelay daemon", daemon.SystemDaemon)
	// if err != nil {
	// 	log.Println("Error: ", err)
	// 	os.Exit(1)
	// }
	// service := &Service{srv}
	// fmt.Println(service.Daemon.Install())
	// fmt.Println("sTART")
	// fmt.Println(service.Daemon.Start())

	// return &os.Process{}, nil

	binary, err := exec.LookPath(os.Args[0])
	if err != nil {
		log.Fatalln("Failed to lookup binary:", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// BUG: when closing the terminal window once detached the process is still killed
	p, err := os.StartProcess(binary, append([]string{binary, "-" + forkIdentifier}, os.Args[1:]...), &os.ProcAttr{Dir: cwd, Env: nil,
		// Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		Files: []*os.File{nil, nil, nil},
		Sys:   nil,
	})
	if err != nil {
		return nil, err
	}

	// cmd := exec.Command(os.Args[0], append([]string{"-" + forkIdentifier}, os.Args[1:]...)...)
	// err := cmd.Start()
	// if err != nil {
	// 	return nil, err
	// }

	// if err := cmd.Process.Release(); err != nil {
	// 	return nil, err
	// }

	return p, nil
}
