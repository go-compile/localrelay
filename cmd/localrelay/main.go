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
		// TODO: add restart which will ask the daemon to fork it self
		case "stop":
			if err := stopDaemon(); err != nil {
				fmt.Println(err)
			}

			fmt.Println("Daemon has been shutdown")
			return
		default:
			fmt.Printf("Unrecognised command %q\n", opt.commands[i])
			return
		}
	}
}
