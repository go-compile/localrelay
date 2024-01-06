package main

import (
	"encoding/json"
	"net/url"

	"github.com/go-compile/localrelay/internal/ipc"
	"github.com/go-compile/localrelay/pkg/api"
	"github.com/pkg/errors"
)

// serviceRun takes paths to relay config files and then connects via IPC to
// instruct the service to run these relays
func serviceRun(relays []string) error {
	c, err := api.Connect()
	if err != nil {
		return err
	}

	defer c.Close()

	for _, relay := range relays {
		r, err := c.StartRelay(relay)

		for _, v := range r {
			Println(v)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func serviceStatus() (*api.Status, error) {
	c, err := api.Connect()
	if err != nil {
		return nil, err
	}

	defer c.Close()

	status, err := c.GetStatus()
	return status, err
}

func stopRelay(relayName string) error {
	client, conn, err := ipc.Connect()
	if err != nil {
		return err
	}

	defer conn.Close()

	resp, err := client.Get("http://lr/stop/" + url.PathEscape(relayName))
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case 200:
		Printf("Relay %q has been stopped.\n", relayName)
	case 500:
		Printf("Failed to stop relay.\n")
	case 404:
		Printf("Relay not found.\n")
	default:
		Printf("Unknown response %d.\n", resp.StatusCode)
	}

	return nil
}

func activeConnections() ([]api.Connection, error) {
	client, conn, err := IPCConnect()
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	resp, err := client.Get("http://lr/connections")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("failed to fetch connections")
	}

	var pool []connection
	if err := json.NewDecoder(resp.Body).Decode(&pool); err != nil {
		return nil, err
	}

	return pool, nil
}

func dropAll() error {
	client, conn, err := IPCConnect()
	if err != nil {
		return err
	}

	defer conn.Close()

	resp, err := client.Get("http://lr/drop")
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case 200:
		Printf("All connections have been dropped.\r\n")
	default:
		Printf("Failed to drop connections. Status code: %d.\r\n", resp.StatusCode)
	}

	return nil
}

func dropIP(ip string) error {
	client, conn, err := IPCConnect()
	if err != nil {
		return err
	}

	defer conn.Close()

	resp, err := client.Get("http://lr/drop/ip/" + url.PathEscape(ip))
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case 200:
		Printf("All connections from %q have been dropped.\r\n", ip)
	default:
		Printf("Failed to drop connections. Status code: %d.\n", resp.StatusCode)
	}

	return nil
}

func dropRelay(relay string) error {
	client, conn, err := IPCConnect()
	if err != nil {
		return err
	}

	defer conn.Close()

	resp, err := client.Get("http://lr/drop/relay/" + url.PathEscape(relay))
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case 200:
		Printf("All connections from %q have been dropped.\r\n", relay)
	default:
		Printf("Failed to drop connections. Status code: %d.\n", resp.StatusCode)
	}

	return nil
}
