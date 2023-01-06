package main

import (
	"testing"
	"time"
)

func TestIPCWindows(t *testing.T) {
	go func() {
		if err := IPCListen(); err != nil {
			// if IPC listen fails make sure your host system isn't
			// already running localrelay. Run localrelay stop
			t.Fatal(err)
		}
	}()

	time.Sleep(time.Millisecond * 50)

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
