package ipc

import (
	"net"
	"net/http"

	"gopkg.in/natefinch/npipe.v2"
)

// ipcConnect will use name pipes to communicate to the daemon
func Connect() (*http.Client, net.Conn, error) {
	conn, err := npipe.DialTimeout(`\\.\pipe\`+serviceName, ipcTimeout)
	if err != nil {
		return nil, nil, err
	}

	httpClient := newHTTPClient(conn)

	return httpClient, conn, nil
}
