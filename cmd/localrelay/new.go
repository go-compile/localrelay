package main

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/chzyer/readline"
	"github.com/go-compile/localrelay/v2"
	"github.com/naoina/toml"
	"github.com/pkg/errors"
)

var (
	// ErrParseBool is returned when a boolean can not be parsed
	ErrParseBool = errors.New("cannot parse boolean")

	relayNameFormat = regexp.MustCompile("^[a-zA-Z0-9]+(?:-[a-zA-Z0-9]+)*$")
)

func newRelay(opt *options, i int, cmd []string) error {
	if len(opt.commands)-1 <= i {
		Println("[WARN] Relay name was not provided.")
		return nil
	}

	if err := createConfigDir(); err != nil {
		Printf("[WARN] Failed to create config dir: %s\n", err)
	}

	name := cmd[i+1]
	if !validateName(name) {
		Println("[WARN] Invalid relay name.")
		return nil
	}

	if opt.host == "" {
		Println("[WARN] Host was not set.")
		return nil
	}

	if opt.destination == "" {
		Println("[WARN] Destination was not set.")
		return nil
	}

	switch strings.ToLower(opt.proxy.Protocol) {
	case "", "socks5":
		// validate socks5 or empty is ok
	default:
		Println("[WARN] Unsupported proxy type.")
		return nil
	}

	listener := localrelay.TargetLink(string(opt.proxyType) + "://" + opt.host)

	relay := Relay{
		Name:     name,
		Listener: listener,
		Logging:  opt.logs,

		Tls: TLS{
			Certificate: opt.certificate,
			Private:     opt.key,
		},

		Loadbalance: Loadbalance{
			Enabled: opt.loadbalance,
		},

		Proxies:     make(map[string]Proxy),
		AutoRestart: !opt.DisableAutoStart,
	}

	// assign the mutliple remotes
	dsts := strings.Split(opt.destination, ",")
	for di, dst := range dsts {
		destination := localrelay.TargetLink(string(opt.proxyType) + "://" + dst)
		if opt.proxy.IsSet() && !contains(opt.proxyIgnore, di+1) {
			destination += "/?proxy=proxy-a"
		}

		relay.Destinations = append(relay.Destinations, destination)
	}

	if opt.proxy.IsSet() {
		relay.Proxies["proxy-a"] = opt.proxy
	}

	filename := name + ".toml"
	// If output file has been set use that instead
	if opt.output != "" {
		filename = opt.output
	} else if opt.store {
		// store config file in daemon config dir
		filename = filepath.Join(relaysDir(), filename)
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

			Printf("[Info] Relay config written to %s\n", filename)

			return nil
		}

		return errors.Wrap(err, "opening file")
	}

	defer f.Close()

	prompt, err := readline.New("> ")
	if err != nil {
		return err
	}

	Println("File already exits, do you want to overwrite it?")
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
		Println("[Info] Aborting, file was not overwritten")
		return nil
	}

	if err := f.Truncate(0); err != nil {
		return err
	}

	if err := toml.NewEncoder(f).Encode(relay); err != nil {
		return err
	}

	Printf("[Info] Relay config written to %s\n", filename)

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

func createConfigDir() error {
	home := configSystemDir()
	dir := filepath.Join(home, configDirSuffix)

	exists, err := pathExists(dir)
	if err != nil {
		return err
	}

	// already exists, don't recreate it
	if !exists {
		if err := os.Mkdir(dir, 0644); err != nil {
			return err
		}
	}

	return createLogDir()
}

func createLogDir() error {
	if runtime.GOOS == "windows" {
		return nil
	}

	dir := "/var/log/localrelay/"

	exists, err := pathExists(dir)
	if err != nil {
		return err
	}

	// already exists, don't recreate it
	if exists {
		return nil
	}

	return os.MkdirAll(dir, 0644)
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	return !os.IsNotExist(err), nil
}

func validateName(name string) bool {
	return relayNameFormat.MatchString(name)
}

func contains(l []int, x int) bool {
	for i := 0; i < len(l); i++ {
		if l[i] == x {
			return true
		}
	}

	return false
}
