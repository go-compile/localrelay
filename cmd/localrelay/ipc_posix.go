//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || !windows
// +build darwin dragonfly freebsd linux netbsd openbsd solaris !windows

package main

import (
	"net"
	"net/http"
	"os"
	"strconv"
	"syscall"
	"time"
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
		Timeout: time.Second * 15,
	}

	return httpClient, conn, nil
}

func fileOwnership(stat os.FileInfo) (string, error) {
	s := stat.Sys().(*syscall.Stat_t)
	uid := s.Uid
	gid := s.Gid

	if uid != gid {
		return strconv.Itoa(int(uid)) + "," + strconv.Itoa(int(gid)), nil
	}

	return strconv.Itoa(int(uid)), nil
}

func runningAsRoot() bool {
	return os.Geteuid() == 1
}
