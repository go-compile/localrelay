package main

import (
	"time"

	"github.com/go-compile/localrelay"
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
	In, Out, Active, DialAvg int
	TotalConns               uint64
}
