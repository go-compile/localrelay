package main

import (
	"testing"
	"time"
)

func TestIPCWindows(t *testing.T) {
	go func() {
		if err := IPCListen(); err != nil {
			t.Fatal(err)
		}
	}()

	time.Sleep(time.Millisecond * 50)

	_, err := serviceStatus()
	if err != nil {
		t.Fatal(err)
	}
}
