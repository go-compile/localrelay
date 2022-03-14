package main

import (
	"fmt"
	"strings"
)

const (
	// VERSION uses semantic versioning
	VERSION = "v0.1.0-alpha"
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
		}
	}
}
