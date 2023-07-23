//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || !windows
// +build darwin dragonfly freebsd linux netbsd openbsd solaris !windows

package main

import (
	"io"
	"net"

	"github.com/pkg/errors"
)

func createTmpIPC(connCh chan net.Conn) (string, io.Closer, error) {
	return "", nil, errors.New("not supported on your platform")
}

func elevatePrivileges(args []string) error {
	return errors.New("not supported on your platform")
}

func fork() error {
	return errors.New("not supported on your platform")
}

func forwardIO(opt *options) (net.Conn, error) {
	return nil, nil
}
