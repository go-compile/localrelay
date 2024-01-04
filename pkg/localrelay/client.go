package localrelay

import (
	"net"
	"net/http"

	"github.com/go-compile/localrelay/internal/ipc"
)

type Client struct {
	conn net.Conn
	hc   *http.Client
}

// Connect establishes a connection to the IPC socket
func Connect() (*Client, error) {
	hc, conn, err := ipc.Connect()
	if err != nil {
		return nil, err
	}

	return &Client{
		conn: conn,
		hc:   hc,
	}, nil
}

// Close disconnects from the IPC socket
func (c *Client) Close() error {
	return c.conn.Close()
}
