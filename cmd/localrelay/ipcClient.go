package main

import (
	"github.com/go-compile/localrelay/pkg/api"
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
	c, err := api.Connect()
	if err != nil {
		return err
	}

	defer c.Close()

	if err := c.StopRelay(relayName); err == nil {
		Printf("Relay %q has been stopped.\n", relayName)
		return nil
	}

	return err
}

func activeConnections() ([]api.Connection, error) {
	c, err := api.Connect()
	if err != nil {
		return nil, err
	}

	defer c.Close()

	return c.GetConnections()
}

func dropAll() error {
	c, err := api.Connect()
	if err != nil {
		return err
	}

	defer c.Close()

	return c.DropAll()
}

func dropIP(ip string) error {
	c, err := api.Connect()
	if err != nil {
		return err
	}

	defer c.Close()

	if err := c.DropIP(ip); err == nil {
		Printf("All connections from %q have been dropped.\r\n", ip)
		return nil
	}

	Printf("Failed to drop connections. Err: %s.\n", err)
	return nil
}

func dropRelay(relay string) error {
	c, err := api.Connect()
	if err != nil {
		return err
	}

	defer c.Close()

	if err := c.DropRelay(relay); err == nil {
		Printf("All connections from %q have been dropped.\r\n", relay)
		return nil
	} else {
		Printf("Failed to drop connections. Status code: %s.\n", err)
	}

	return nil
}
