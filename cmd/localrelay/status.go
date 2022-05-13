package main

import (
	"fmt"
	"log"

	"github.com/containerd/console"
)

func relayStatus() error {

	s, err := getDaemonStatus()
	if err != nil {
		fmt.Printf("Daemon:    \x1b[31m [OFFLINE] \x1b[0m\r\n")

		return nil
	}

	// make terminal raw to allow the use of colour on windows terminals
	current := console.Current()
	defer current.Reset()

	if err := current.SetRaw(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\r\nDaemon:      \x1b[102m\x1b[30m [RUNNING] \x1b[0m\r\n")
	fmt.Printf("PID:         [%d]\r\n", s.Pid)
	fmt.Printf("Version:     [%s]\r\n", s.Version)
	fmt.Printf("Relays:      [%d]\r\n", len(s.Relays))

	totalConns := 0
	in := 0
	out := 0
	active := 0

	for _, m := range s.Metrics {
		totalConns += int(m.TotalConns)
		in += int(m.In)
		out += int(m.In)
		active += int(m.Active)
	}

	fmt.Printf("Total Conns: [%d]\r\n", totalConns)
	fmt.Printf("Active:      [%d]\r\n", active)
	fmt.Printf("In/Out:      [%d/%d]\r\n", in, out)

	for i := range s.Relays {
		fmt.Printf("  \x1b[90m%.2d\x1b[0m: %s\r\n      %s -> %s\r\n", i+1, s.Relays[i].Name, s.Relays[i].Host, s.Relays[i].ForwardAddr)
	}

	return nil
}
