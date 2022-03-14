package main

import (
	"fmt"
	"os"

	"github.com/naoina/toml"
	"github.com/pkg/errors"
)

func runRelays(opt *options, i int, cmd []string) error {

	// Read all relay config files and decode them
	relays := make([]Relay, 0, len(cmd[i+1:]))
	for _, file := range cmd[i+1:] {

		f, err := os.Open(file)
		if err != nil {
			return errors.Wrapf(err, "file:%q", file)
		}

		var relay Relay
		if err := toml.NewDecoder(f).Decode(&relay); err != nil {
			f.Close()
			return err
		}

		relays = append(relays, relay)

		f.Close()
	}

	if len(relays) == 0 {
		fmt.Println("[WARN] No relay configs provided.")
		return nil
	}

	fmt.Println(relays)

	return nil
}
