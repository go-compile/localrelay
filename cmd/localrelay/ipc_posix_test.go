//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || !windows
// +build darwin dragonfly freebsd linux netbsd openbsd solaris !windows

package main

import (
	"testing"
	"time"
)

func TestIPCPosix(t *testing.T) {

	ipcPathPrefix = "./"
	go func() {
		if err := IPCListen(); err != nil {
			t.Fatal(err)
		}
	}()

	time.Sleep(time.Second)

	_, err := serviceStatus()
	if err != nil {
		t.Fatal(err)
	}

	// close IPC listener
	ipcListener.Close()
}
