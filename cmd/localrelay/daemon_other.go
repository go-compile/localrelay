//go:build !windows
// +build !windows

package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
)

var (
	daemonNotSupported = "Daemon is not supported on your platform."
)

func launchDaemon() {
	fmt.Println(daemonNotSupported)
	os.Exit(0)
}

func getDaemonStatus() (*status, error) {
	fmt.Println(daemonNotSupported)
	os.Exit(0)

	return nil, nil
}

func stopDaemon() error {
	return errors.New(daemonNotSupported)
}
