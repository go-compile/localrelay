//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || !windows
// +build darwin dragonfly freebsd linux netbsd openbsd solaris !windows

package main

import (
	"io"
	"net"
)

var (
	// ipcPathPrefix is the dir which comes before the unix socket
	ipcPathPrefix = "/var/run/"
)

func IPCConnect() (io.ReadWriteCloser, error) {
	return net.DialTimeout("unix", ipcPathPrefix+ipcSocket, ipcTimeout)
}

func IPCListen() error {

	l, err := net.Listen("unix", ipcPathPrefix+ipcSocket)
	if err != nil {
		return err
	}
	ipcListener = l

	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			if err == net.ErrClosed {
				return err
			}

			continue
		}

		go handleConn(conn, l)
	}
}
