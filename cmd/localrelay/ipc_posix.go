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

func IPCConnect() (*http.Client, net.Conn, error) {
	conn, err := net.DialTimeout("unix", ipcPathPrefix+ipcSocket, ipcTimeout)
	if err != nil {
		return nil, nil, err
	}

	// make a http client which always uses the socket.
	// When making a HTTP request provide any host, it does not need to exist.
	//
	// Example:
	//  http://lr/status
	httpClient := &http.Client{
		Transport: &http.Transport{Dial: func(network, addr string) (net.Conn, error) {
			return conn, nil
		}},
	}

	return httpClient, conn, nil
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
