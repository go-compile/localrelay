package main

import (
	"fmt"
	"strings"

	"github.com/go-compile/localrelay"
)

const (
	// VERSION uses semantic versioning
	VERSION = localrelay.VERSION
)

func main() {
	opt, err := parseArgs()
	if err == nil && opt == nil {
		return
	}

	if err != nil {
		fmt.Println(err)
		return
	}

	if len(opt.commands) == 0 {
		help()
		return
	}

	for i := 0; i < len(opt.commands); i++ {

		switch strings.ToLower(opt.commands[i]) {
		case "help", "h", "?":
			help()
			return
		case "version", "v":
			version()
			return
		case "new":
			if err := newRelay(opt, i, opt.commands); err != nil {
				fmt.Println(err)
			}

			return
		case "run":
			if err := runRelays(opt, i, opt.commands); err != nil {
				fmt.Println(err)
			}

			return
		case "stop":
			if err := stopDaemon(); err != nil {
				fmt.Println(err)
			}

			fmt.Println("Daemon has been shutdown")
			return
		case "restart":
			if err := relayStatus(); err != nil {
				fmt.Println(err)
				return
			}

			if err := forkDeamon(); err != nil {
				fmt.Println("Could not restart/fork daemon")
				return
			}

			fmt.Println("Daemon has been restarted")
			return
		case "status":
			if err := relayStatus(); err != nil {
				fmt.Println(err)
			}

			return
			// TODO: add logs. Connect via IPC and show live view
			// TODO: add metrics with -interval=5s
			// For attached and detached instances
			// TODO: add autoload home dir config argument, use in addition of run configs
		default:
			fmt.Printf("Unrecognised command %q\n", opt.commands[i])
			return
		}
	}
}
