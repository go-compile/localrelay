package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/naoina/toml"
	"github.com/pkg/errors"
)

var (
	// ErrParseBool is returned when a boolean can not be parsed
	ErrParseBool = errors.New("cannot parse boolean")
)

func newRelay(opt *options, i int, cmd []string) error {
	if len(opt.commands)-1 <= i {
		fmt.Println("[WARN] Relay name was not provided.")
		return nil
	}

	name := cmd[i+1]

	if opt.host == "" {
		fmt.Println("[WARN] Host was not set.")
		return nil
	}

	if opt.destination == "" {
		fmt.Println("[WARN] Destination was not set.")
		return nil
	}

	// TODO: add TLS certificate options
	// TODO: add logging options
	relay := Relay{
		Name:        name,
		Host:        opt.host,
		Destination: opt.destination,
		Kind:        opt.proxyType,
		Logging:     "stdout",
		ProxyIgnore: opt.proxyIgnore,

		Proxy: &opt.proxy,
	}

	filename := name + ".toml"
	// If output file has been set use that instead
	if opt.output != "" {
		filename = opt.output
	}

	f, err := os.OpenFile(filename, os.O_WRONLY, os.ModeExclusive)
	if err != nil {
		if os.IsNotExist(err) {

			f, err := os.Create(filename)
			if err != nil {
				return err
			}

			if err := toml.NewEncoder(f).Encode(relay); err != nil {
				return err
			}

			if err := f.Close(); err != nil {
				return err
			}

			fmt.Printf("[Info] Relay config written to %s\n", filename)

			return nil
		}

		return errors.Wrap(err, "opening file")
	}

	defer f.Close()

	prompt, err := readline.New("> ")
	if err != nil {
		return err
	}

	fmt.Println("File already exits, do you want to overwrite it?")
	prompt.SetPrompt("Overwrite (y/n): ")
	overwrite, err := prompt.ReadlineWithDefault("n")
	if err != nil {
		return err
	}

	ow, err := parseBool(overwrite)
	if err != nil {
		return err
	}

	if !ow {
		fmt.Println("[Info] Aborting, file was not overwritten")
		return nil
	}

	if err := f.Truncate(0); err != nil {
		return err
	}

	if err := toml.NewEncoder(f).Encode(relay); err != nil {
		return err
	}

	fmt.Printf("[Info] Relay config written to %s\n", filename)

	return nil
}

func parseBool(input string) (bool, error) {
	switch strings.ToLower(input) {
	case "true", "1", "yes", "on", "active", "y":
		return true, nil
	case "false", "0", "no", "off", "disabled", "n":
		return false, nil
	default:
		return false, ErrParseBool
	}
}
