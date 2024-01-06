package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-compile/localrelay/internal/ipc"
)

var (
	ErrNotOk    = errors.New("status code not ok")
	ErrFailure  = errors.New("localrelay failed executing the requested action")
	ErrNotFound = errors.New("relay not found")
)

type Client struct {
	conn net.Conn
	hc   *http.Client
}

type msgResponse struct {
	Message string `json:"message"`
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

	return nil
}

func (c *Client) DropAll() error {
	resp, err := c.hc.Get("http://lr/drop")
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return ErrNotOk
	}

	return nil
}

func (c *Client) StopRelay(relay string) error {
	resp, err := c.hc.Get("http://lr/stop/relay/" + url.PathEscape(relay))
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case 200:
		return nil
	case 500:
		return ErrFailure
	case 404:
		return ErrNotFound
	default:
		return errors.New("unknown respose code")
	}
}

func (c *Client) StartRelay(relays ...string) (responses []string, err error) {
	for _, relay := range relays {
		// make post request to run relay. Use strconv instead of json encoding for performance
		resp, err := c.hc.Post("http://lr/run", "application/json", bytes.NewBuffer([]byte("["+strconv.Quote(relay)+"]")))
		if err != nil {
			return responses, err
		}

		var response msgResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return responses, err
		}

		responses = append(responses, response.Message)

		switch resp.StatusCode {
		case 404:
			return responses, ErrNotFound
		case 500:
			return responses, ErrFailure
		}
	}

	return responses, nil
}
