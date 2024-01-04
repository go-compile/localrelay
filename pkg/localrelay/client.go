package localrelay

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"

	"github.com/go-compile/localrelay/internal/ipc"
)

var (
	ErrNotOk = errors.New("status code not ok")
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

func (c *Client) GetStatus() (*Status, error) {
	resp, err := c.hc.Get("http://lr/status")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, ErrNotOk
	}

	var status Status
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, err
	}

	return &status, nil
}

func (c *Client) GetConnections() ([]Connection, error) {
	resp, err := c.hc.Get("http://lr/connections")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, ErrNotOk
	}

	var pool []Connection
	if err := json.NewDecoder(resp.Body).Decode(&pool); err != nil {
		return nil, err
	}

	return pool, nil
}

func (c *Client) DropRelay(relay string) error {
	resp, err := c.hc.Get("http://lr/drop/relay/" + url.PathEscape(relay))
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return ErrNotOk
	}

	if resp.StatusCode != 200 {
		return ErrNotOk
	}

	return nil
}

func (c *Client) DropIP(ip string) error {
	resp, err := c.hc.Get("http://lr/drop/ip/" + url.PathEscape(ip))
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return ErrNotOk
	}

	if resp.StatusCode != 200 {
		return ErrNotOk
	}

	return nil
}
