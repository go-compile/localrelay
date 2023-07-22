package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/containerd/console"
)

func displayOpenConns() error {
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

	for _, conn := range conns {
		fmt.Printf("%s -> %s (%s) (%s)\r\n", conn.RemoteAddr, conn.ForwardedAddr, conn.RelayName, formatDuration(time.Since(time.Unix(conn.Opened, 0))))
	}

	return nil
}
