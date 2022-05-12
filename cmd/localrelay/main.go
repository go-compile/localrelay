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
		case "run":
			if err := runRelays(opt, i, opt.commands); err != nil {
				fmt.Println(err)
			}
		default:
			fmt.Printf("Unrecognised command %q\n", opt.commands[i])
			return
		}
	}
}
