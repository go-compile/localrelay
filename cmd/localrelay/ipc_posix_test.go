//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || !windows
// +build darwin dragonfly freebsd linux netbsd openbsd solaris !windows

package main

import (
	"testing"
	"time"

	"github.com/go-compile/localrelay/internal/ipc"
)

func TestIPCPosix(t *testing.T) {
	go func() {
		l, err := ipc.NewListener()
		if err != nil {
			t.Fatal(err)
		}

		defer l.Close()

		ipcListener = l

		err = ipc.ListenServe(l, newIPCServer())
		if err != nil {
			t.Fatal(err)
		}
	}()

	time.Sleep(time.Second)

	// due to the above embeded function being in a different
	// gorutine, t.Fatal will only effect the above subroutine.
	// Hence needing to check if ipcListener is nil.
	if ipcListener == nil {
		t.Fatal("ipc listener could not startup")
	}

	_, err := serviceStatus()
	if err != nil {
		t.Fatal(err)
	}

	// close IPC listener
	ipcListener.Close()
}
