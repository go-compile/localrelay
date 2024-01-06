package main

import (
	"net"
	"net/http"
	"os"
	"time"

	"golang.org/x/sys/windows"
	"gopkg.in/natefinch/npipe.v2"
)

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

func fileOwnership(stat os.FileInfo) (string, error) {
	// TODO: get owner of file on windows
	return "", nil
}

func runningAsRoot() bool {
	token := windows.GetCurrentProcessToken()
	defer token.Close()

	return token.IsElevated()
}
