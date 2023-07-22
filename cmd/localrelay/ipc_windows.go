package main

import (
	"net"
	"net/http"
	"time"

	"gopkg.in/natefinch/npipe.v2"
)

// IPCListen for windows uses name pipes to communicate
func IPCListen() error {

	l, err := npipe.Listen(`\\.\pipe\` + serviceName)
	if err != nil {
		return err
	}

	ipcListener = l

	defer l.Close()

	srv := newIPCServer()

	for {
		conn, err := l.Accept()
		if err != nil {
			if err == net.ErrClosed {
				return err
			}

			continue
		}

		go handleConn(conn, srv, l)
	}
}

// IPCConnect will use name pipes to communicate to the daemon
func IPCConnect() (*http.Client, net.Conn, error) {
	conn, err := npipe.DialTimeout(`\\.\pipe\`+serviceName, ipcTimeout)
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
		Timeout: time.Second * 15,
	}

	return httpClient, conn, nil
}
