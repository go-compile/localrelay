package main

import (
	"io"
	"net"

	"github.com/valyala/fasthttp"
)

type daemon struct{}

const (
	serviceName        = "localrelayd"
	ipcSocket          = "localrelay.ipc.socket"
	serviceDescription = "Localrelay daemon relay runner"
)

var (
	ipcListener io.Closer
)

// handleConn takes a conn and handles each command
func handleConn(conn net.Conn, srv *fasthttp.Server, l io.Closer) {
	defer conn.Close()

	srv.ServeConn(conn)
}
