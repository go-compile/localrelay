package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/containerd/console"
)

func displayOpenConns(opt *options, onlyIPS bool) error {
	// make terminal raw to allow the use of colour on windows terminals
	current, _ := console.ConsoleFromFile(os.Stdout)
	// NOTE: Docker healthchecks will panic "provided file is not a console"

	if current != nil {
		defer current.Reset()
	}

	if current != nil {
		if err := current.SetRaw(); err != nil {
			log.Println(err)
		}
	}

	// we don't set terminal to raw here because print statements don't use
	// carriage returns
	conns, err := activeConnections()
	if err != nil {

		fmt.Printf("Daemon:    \x1b[31m [OFFLINE] \x1b[0m\r\n")
		fmt.Println(err)

		// exit with error
		os.Exit(1)
	}

	filteredRelays := []string{}

	// build filter list for relays
	if len(opt.commands) > 1 {
		for _, relayName := range opt.commands[1:] {
			if !validateName(relayName) {
				fmt.Println("[WARN] Invalid relay name.")
				return nil
			}

			filteredRelays = append(filteredRelays, relayName)
		}
	}

	for _, conn := range conns {
		if len(filteredRelays) != 0 {
			if arrayContains(filteredRelays, conn.RelayName) {
				printConn(conn, onlyIPS)
			}
		} else {
			printConn(conn, onlyIPS)
		}
	}

	return nil
}

func printConn(conn connection, onlyIPS bool) {
	if onlyIPS {
		fmt.Printf("%s\r\n", conn.RemoteAddr)
		return
	}

	fmt.Printf("%s -> %s (%s) (%s)\r\n", conn.RemoteAddr, conn.ForwardedAddr, conn.RelayName, formatDuration(time.Since(time.Unix(conn.Opened, 0))))
}

func arrayContains(arr []string, element string) bool {
	for i := 0; i < len(arr); i++ {
		if arr[i] == element {
			return true
		}
	}

	return false
}

func dropConns(opt *options) error {
	// make terminal raw to allow the use of colour on windows terminals
	current, _ := console.ConsoleFromFile(os.Stdout)
	// NOTE: Docker healthchecks will panic "provided file is not a console"

	if current != nil {
		defer current.Reset()
	}

	if current != nil {
		if err := current.SetRaw(); err != nil {
			log.Println(err)
		}
	}

	return dropAll()
}

func dropConnsIP(opt *options) error {
	// make terminal raw to allow the use of colour on windows terminals
	current, _ := console.ConsoleFromFile(os.Stdout)
	// NOTE: Docker healthchecks will panic "provided file is not a console"

	if current != nil {
		defer current.Reset()
	}

	if current != nil {
		if err := current.SetRaw(); err != nil {
			log.Println(err)
		}
	}

	if len(opt.commands) < 2 {
		fmt.Println("Provide an ip address.")
		return nil
	}

	return dropIP(opt.commands[1])
}

func dropConnsRelay(opt *options) error {
	// make terminal raw to allow the use of colour on windows terminals
	current, _ := console.ConsoleFromFile(os.Stdout)
	// NOTE: Docker healthchecks will panic "provided file is not a console"

	if current != nil {
		defer current.Reset()
	}

	if current != nil {
		if err := current.SetRaw(); err != nil {
			log.Println(err)
		}
	}

	if len(opt.commands) < 2 {
		fmt.Println("Provide a relay name.")
		return nil
	}

	return dropRelay(opt.commands[1])
}
