package main

import (
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/go-compile/localrelay/v2"
	"github.com/pkg/errors"
)

var (
	// ErrFailedCheckUpdate is returned when the latest version could not be fetched
	ErrFailedCheckUpdate = errors.New("failed to check for updates")
)

type options struct {
	host        string
	destination string
	proxyType   localrelay.ProxyType
	proxy       Proxy
	output      string
	proxyIgnore []int
	logs        string

	certificate string
	key         string

	commands []string
	detach   bool

	isFork           bool
	DisableAutoStart bool
	store            bool

	ipcPipe string

	interval time.Duration
}

/*
localrelay new nextcloud -host 127.0.0.1:87 -destination example.com -tcp -proxy socks5://127.0.0.1:9050

localrelay run ./nextcloud.toml ./git.toml

> Overwrite options such as logs for all profiles
localrelay run ./nextcloud.toml ./git.toml -metrics=5s -logs=./relay.log
*/
func parseArgs() (*options, error) {
	args := os.Args[1:]

	opt := &options{
		logs:      "stdout",
		interval:  time.Second,
		proxyType: localrelay.ProxyTCP,
	}

	for i := 0; i < len(args); i++ {
		if !strings.HasPrefix(args[i], "-") {
			opt.commands = append(opt.commands, args[i])
			continue
		}

		// Strip prefix
		arg := strings.SplitN(args[i][1:], "=", 2)

		switch strings.ToLower(arg[0]) {
		case forkIdentifier:
			opt.isFork = true
		case "version":
			version()
			return nil, nil
		case "interval", "refresh":
			value, err := getAnswer(args, arg, &i)
			if err != nil {
				return nil, err
			}

			dur, err := time.ParseDuration(value)
			if err != nil {
				return nil, err
			}

			opt.interval = dur
		case "host", "lhost":
			value, err := getAnswer(args, arg, &i)
			if err != nil {
				return nil, err
			}

			opt.host = value
		case "log", "logs":
			value, err := getAnswer(args, arg, &i)
			if err != nil {
				return nil, err
			}

			opt.logs = value
		case "cert", "certificate":
			value, err := getAnswer(args, arg, &i)
			if err != nil {
				return nil, err
			}

			opt.certificate = value
		case "key":
			value, err := getAnswer(args, arg, &i)
			if err != nil {
				return nil, err
			}

			opt.key = value
		case "disable_autostart", "disable_auto_start", "nostart", "noauto":
			opt.DisableAutoStart = true
		case "ipc-stream-io-pipe":
			value, err := getAnswer(args, arg, &i)
			if err != nil {
				return nil, err
			}

			opt.ipcPipe = value
		case "detach", "bg", "d":
			opt.detach = true
		case "store", "s":
			opt.store = true
		case "timeout":
			value, err := getAnswer(args, arg, &i)
			if err != nil {
				return nil, err
			}

			dur, err := time.ParseDuration(value)
			if err != nil {
				return nil, err
			}

			Printf("Timeout set to: %dms\n", dur.Milliseconds())
			localrelay.Timeout = dur
		case "destination", "dst", "rhost":
			value, err := getAnswer(args, arg, &i)
			if err != nil {
				return nil, err
			}

			opt.destination = value
		case "output", "o":
			value, err := getAnswer(args, arg, &i)
			if err != nil {
				return nil, err
			}

			opt.output = value
		case "proxyignore", "proxy_ignore":
			value, err := getAnswer(args, arg, &i)
			if err != nil {
				return nil, err
			}

			var ignored []int
			for _, index := range strings.Split(value, ",") {
				i, err := strconv.Atoi(index)
				if err != nil {
					return nil, err
				}

				ignored = append(ignored, i)
			}

			opt.proxyIgnore = ignored
		case "proxy":
			value, err := getAnswer(args, arg, &i)
			if err != nil {
				return nil, err
			}

			// Parse proxy url
			// socks5://127.0.0.1:9050
			prox, err := url.Parse(value)
			if err != nil {
				return nil, err
			}

			pw, _ := prox.User.Password()

			opt.proxy.Protocol = prox.Scheme
			opt.proxy.Address = prox.Host
			opt.proxy.Username = prox.User.Username()
			opt.proxy.Password = pw
		case "tcp":
			opt.proxyType = localrelay.ProxyTCP
		case "udp":
			opt.proxyType = localrelay.ProxyUDP
		case "http":
			opt.proxyType = localrelay.ProxyHTTP
		case "https":
			opt.proxyType = localrelay.ProxyHTTPS
		case "help", "h", "?":
			help()
			if len(os.Args) >= 3 {
				Println("\n\n[Warn] It looks like you accidentally used -h instead of -host")
			}
			return nil, nil
		default:
			Printf("Unknown argument %q\n", arg)
			return nil, nil
		}

	}

	return opt, nil
}

func getAnswer(args []string, arg []string, i *int) (string, error) {
	// If arg is a KV pair
	if len(arg) == 2 {
		return arg[1], nil
	}

	// Check if there are any more key values
	if len(args)-1 <= *i {
		return "", errors.New("Expected value to be paired with argument")
	}

	// Skip next argument as we are going to use it now
	*i++

	// Check if next value is a argument
	if x := args[*i]; len(x) > 0 && x[0] == '-' {
		return "", errors.New("A value can not be a argument")
	}

	return args[*i], nil
}

func help() {
	Printf("LocalRelay CLI - %s\n", VERSION)
	Println()
	Println("Usage:")
	Println("  localrelay new <relay_name> -host 127.0.0.1:8080 -destination example.com:80")
	Println("    -output=<file_location> -tcp -http -https -proxy socks5://127.0.0.1:9050")
	Println()
	Println("  localrelay run <relay_config>")
	Println("  localrelay run <relay_config> -detach")
	Println("  localrelay run <relay_config> <relay_config2>...")
	Println()
	Println("  localrelay start")
	Println("  localrelay status")
	Println("  localrelay monitor")
	Println("  localrelay connections")
	Println("  localrelay connections <relay>")
	Println("  localrelay ips")
	Println("  localrelay drop")
	Println("  localrelay dropip <ip>")
	Println("  localrelay droprelay <relay>")
	Println("  localrelay stop")
	Println("  localrelay stop <relay>")
	Println("  localrelay restart")
	Println("  localrelay install")
	Println("  localrelay uninstall")
	Println()
	Println("Arguments:")
	Printf("  %-28s %s\n", "-host, -lhost", "Set listen host")
	Printf("  %-28s %s\n", "-destination, -dst, -rhost", "Set forward address")
	Printf("  %-28s %s\n", "-tcp", "Set relay to TCP relay")
	Printf("  %-28s %s\n", "-udp", "Set relay to UDP relay")
	Printf("  %-28s %s\n", "-http", "Set relay to HTTP relay")
	Printf("  %-28s %s\n", "-https", "Set relay to HTTPS relay")
	Printf("  %-28s %s\n", "-failover", "Set relay to TCP failover relay")
	Printf("  %-28s %s\n", "-proxy", "Set socks5 proxy via URL")
	Printf("  %-28s %s\n", "-output, -o", "Set output file path")
	Printf("  %-28s %s\n", "-proxy_ignore", "Destination indexes to ignore proxy settings")
	Printf("  %-28s %s\n", "-version", "View version page")
	Printf("  %-28s %s\n", "-timeout", "Set dial timeout for non proxied relays")
	Printf("  %-28s %s\n", "-detach", "Run relay service in background")
	Printf("  %-28s %s\n", "-log", "Specify the file to write logs to")
	Printf("  %-28s %s\n", "-cert", "Set TLS certificate file")
	Printf("  %-28s %s\n", "-key", "Set TLS key file")
	Printf("  %-28s %s\n", "-noauto", "Set relay to not autostart with daemon")
	Printf("  %-28s %s\n", "-store", "Output relay configs to config dir")
	Printf("  %-28s %s\n", "-interval", "Metrics refresh interval")
}

func version() {
	Printf("LocalRelay CLI - %s (%s.%s) [%s]\n", VERSION, BRANCH, COMMIT, runtime.Version())
	Println()
	Println(" A reverse proxying program to allow services e.g. Nextcloud, Bitwarden etc to\n" +
		" be accessed over Tor (SOCKS5) even when the client app do not support\n" +
		" SOCKS proxies.")
	Println()
	Println()
	Println(" github.com/go-compile/localrelay")

	checkForUpdates()
}
