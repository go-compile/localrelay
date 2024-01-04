package ipc

import (
	"io"
	"net"
	"net/http"
	"time"

	"github.com/valyala/fasthttp"
)

const (
	serviceName        = "localrelayd"
	ipcSocket          = "localrelay.ipc.socket"
	serviceDescription = "Localrelay daemon relay runner"
	ipcTimeout         = time.Second
)

// handleConn takes a conn and handles each command
func handleConn(conn net.Conn, srv *fasthttp.Server, l io.Closer) {
	defer conn.Close()
	srv.ServeConn(conn)
}

// newHTTPClient makes a http client which always uses the socket.
// When making a HTTP request provide any host, it does not need to exist.
//
// Example:
//
//	http://lr/status
func newHTTPClient(conn net.Conn) *http.Client {
	return &http.Client{
		Transport: &http.Transport{Dial: func(network, addr string) (net.Conn, error) {
			return conn, nil
		}},
		Timeout: time.Second * 15,
	}
}

func ListenServe(l net.Listener, srv *fasthttp.Server) error {
	defer l.Close()

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
