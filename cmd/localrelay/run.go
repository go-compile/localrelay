package main

import (
	"fmt"
	"os"

	"github.com/naoina/toml"
)

func runRelays(opt *options, i int, cmd []string) error {

	// Read all relay config files and decode them
	relays := make([]Relay, 0, len(cmd[1:]))
	for _, file := range cmd[i:] {

		f, err := os.Open(file)
		if err != nil {
			return err
		}

		var relay Relay
		if err := toml.NewDecoder(f).Decode(&relay); err != nil {
			f.Close()
			return err
		}

		relays = append(relays, relay)

		f.Close()
	}

	fmt.Println(relays)

	return nil
}
