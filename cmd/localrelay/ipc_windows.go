package main

import (
	"io"
	"net"

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

// IPCConnect will use name pipes to communicate to the daemon
func IPCConnect() (io.ReadWriteCloser, error) {
	conn, err := npipe.DialTimeout(`\\.\pipe\`+serviceName, ipcTimeout)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
