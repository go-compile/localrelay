package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-compile/localrelay"
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

	commands []string
	detach   bool

	isFork bool
}

/*
localrelay new nextcloud -host 127.0.0.1:87 -destination example.com -tcp -proxy socks5://127.0.0.1:9050

localrelay run ./nextcloud.toml ./git.toml

> Overwrite options such as logs for all profiles
localrelay run ./nextcloud.toml ./git.toml -metrics=5s -logs=./relay.log
*/
func parseArgs() (*options, error) {
	args := os.Args[1:]

	opt := &options{}

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
		case "host", "lhost":
			value, err := getAnswer(args, arg, &i)
			if err != nil {
				return nil, err
			}

			opt.host = value
		case "detach", "bg":
			opt.detach = true
		case "timeout":
			value, err := getAnswer(args, arg, &i)
			if err != nil {
				return nil, err
			}

			dur, err := time.ParseDuration(value)
			if err != nil {
				return nil, err
			}

			fmt.Printf("Timeout set to: %dms\n", dur.Milliseconds())
			localrelay.Timeout = dur
		case "destination", "d", "dst", "rhost":
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
			opt.proxy.Host = prox.Host
			opt.proxy.Username = prox.User.Username()
			opt.proxy.Password = pw
		case "tcp":
			opt.proxyType = localrelay.ProxyTCP
		case "http":
			opt.proxyType = localrelay.ProxyHTTP
		case "https":
			opt.proxyType = localrelay.ProxyHTTPS
		case "failover", "failovertcp", "failover-tcp", "tcp-failover":
			opt.proxyType = localrelay.ProxyFailOverTCP
		case "help", "h", "?":
			help()
			return nil, nil
		default:
			fmt.Printf("Unknown argument %q\n", arg)
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
	fmt.Printf("LocalRelay CLI - %s\n", VERSION)
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  localrelay new <relay_name> -host 127.0.0.1:8080 -destination example.com:80")
	fmt.Println("    -output=<file_location> -tcp -http -https -proxy socks5://127.0.0.1:9050")
	fmt.Println()
	fmt.Println("  localrelay run <relay_config>")
	fmt.Println("  localrelay run <relay_config> -detach")
	fmt.Println("  localrelay run <relay_config> <relay_config2>...")
	fmt.Println()
	fmt.Println("  localrelay status")
	fmt.Println("  localrelay stop")
	fmt.Println("  localrelay restart")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Printf("  %-28s %s\n", "-host, -lhost", "Set listen host")
	fmt.Printf("  %-28s %s\n", "-destination, -dst, -rhost", "Set forward address")
	fmt.Printf("  %-28s %s\n", "-tcp", "Set relay to TCP relay")
	fmt.Printf("  %-28s %s\n", "-http", "Set relay to HTTP relay")
	fmt.Printf("  %-28s %s\n", "-https", "Set relay to HTTPS relay")
	fmt.Printf("  %-28s %s\n", "-failover", "Set relay to TCP failover relay")
	fmt.Printf("  %-28s %s\n", "-proxy", "Set socks5 proxy via URL")
	fmt.Printf("  %-28s %s\n", "-output, -o", "Set output file path")
	fmt.Printf("  %-28s %s\n", "-proxy_ignore", "Destination indexes to ignore proxy settings")
	fmt.Printf("  %-28s %s\n", "-version", "View version page")
	fmt.Printf("  %-28s %s\n", "-timeout", "Set dial timeout for non proxied relays")
	fmt.Printf("  %-28s %s\n", "-detach", "Run relay service in background")
}

func version() {
	fmt.Printf("LocalRelay CLI - %s\n", VERSION)
	fmt.Println()
	fmt.Println(" A reverse proxying program to allow services e.g. Nextcloud, Bitwarden etc to be accessed\n" +
		" over Tor (SOCKS5) even when the client app do not support SOCKS proxies.")
	fmt.Println()
	fmt.Println()
	fmt.Println(" github.com/go-compile/localrelay")

	checkForUpdates()
}
