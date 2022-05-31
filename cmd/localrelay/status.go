package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/containerd/console"
)

func relayStatus() error {

	// we don't set terminal to raw here because print statements don't use
	// carriage returns
	s, err := serviceStatus()
	if err != nil {

		// make terminal raw to allow the use of colour on windows terminals
		current := console.Current()

		if err := current.SetRaw(); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Daemon:    \x1b[31m [OFFLINE] \x1b[0m\r\n")
		fmt.Println(err)
		current.Reset()

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
	totalRequests := 0
	in := 0
	out := 0
	active := 0

	for _, m := range s.Metrics {
		totalConns += int(m.TotalConns)
		totalRequests += int(m.TotalRequests)
		in += int(m.In)
		out += int(m.In)
		active += int(m.Active)
	}

	fmt.Println("\r")
	fmt.Printf("Total Conns: [%d] Total Requests: [%d]\r\n", totalConns, totalRequests)
	fmt.Printf("Active:      [%d]\r\n", active)
	fmt.Printf("In/Out:      [%s/%s]\r\n", formatBytes(in), formatBytes(out))
	fmt.Printf("Uptime:      [%d minutes]\r\n", time.Unix(s.Started, 0).Minute())
	// BUG: uptime incorrect

	for i := range s.Relays {
		fmt.Printf("  \x1b[90m%.2d\x1b[0m: %s\r\n      %s -> %s\r\n", i+1, s.Relays[i].Name, s.Relays[i].Host, s.Relays[i].ForwardAddr)
	}

	return nil
}

func formatBytes(bytes int) string {
	if unit := 1000; bytes < unit {
		return strconv.Itoa(bytes) + "bytes"
	}

	if unit := 1000000; bytes < unit {
		return strconv.FormatFloat(float64(bytes)/1000, 'f', 2, 64) + "kb"
	}

	if unit := 1000000000; bytes < unit {
		return strconv.FormatFloat(float64(bytes)/1000000, 'f', 2, 64) + "mb"
	}

	return strconv.FormatFloat(float64(bytes)/1000000000, 'f', 2, 64) + "gb"

}
