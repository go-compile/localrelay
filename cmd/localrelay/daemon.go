package main

import (
	"time"

	"github.com/go-compile/localrelay/v2"
	"github.com/pkg/errors"
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
	In, Out, Active, DialAvg  int
	TotalConns, TotalRequests uint64
}

type connection struct {
	LocalAddr  string
	RemoteAddr string
	Network    string

	RelayName string
	RelayHost string

	ForwardedAddr string

	// Opened is a unix timestamp
	Opened int64
}
