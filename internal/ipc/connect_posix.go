//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || !windows
// +build darwin dragonfly freebsd linux netbsd openbsd solaris !windows

package ipc

import (
	"net"
	"net/http"
)

var (
	// ipcPathPrefix is the dir which comes before the unix socket
	ipcPathPrefix = "/var/run/"
)

// ipcConnect will use name pipes to communicate to the daemon
func Connect() (*http.Client, net.Conn, error) {
	conn, err := net.DialTimeout("unix", ipcPathPrefix+ipcSocket, ipcTimeout)
	if err != nil {
		return nil, nil, err
	}

	httpClient := newHTTPClient(conn)

	return httpClient, conn, nil
}

// SetPathPrefix will determin where the IPC listener binds and connects
func SetPathPrefix(prefix string) {
	ipcPathPrefix = prefix
}
