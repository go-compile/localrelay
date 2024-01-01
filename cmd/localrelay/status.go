package main

import (
	"log"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/containerd/console"
	"github.com/go-compile/localrelay/v2"
)

func relayStatus() error {
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
	s, err := serviceStatus()
	if err != nil {
		Printf("Daemon:    \x1b[31m [OFFLINE] \x1b[0m\r\n")
		Println(err)

		// exit with error
		os.Exit(1)
	}

	Printf("\r\nDaemon:      \x1b[102m\x1b[30m [RUNNING] \x1b[0m\r\n")
	Printf("PID:         [%d]\r\n", s.Pid)
	Printf("Version:     [%s]\r\n", s.Version)
	Printf("Relays:      [%d]\r\n", len(s.Relays))

	totalConns := 0
	totalRequests := 0
	in := 0
	out := 0
	active := 0

	for _, m := range s.Metrics {
		totalConns += int(m.TotalConns)
		totalRequests += int(m.TotalRequests)
		in += int(m.In)
		out += int(m.Out)
		active += int(m.Active)
	}

	Println("\r")
	Printf("Total Conns: [%d] Total Requests: [%d]\r\n", totalConns, totalRequests)
	Printf("Active:      [%d]\r\n", active)
	Printf("In/Out:      [%s/%s]\r\n", formatBytes(in), formatBytes(out))
	Printf("Uptime:      [%s]\r\n", formatDuration(time.Since(time.Unix(s.Started, 0))))

	// sort alphabetically
	sort.SliceStable(s.Relays, func(i, j int) bool {
		return s.Relays[i].Name < s.Relays[j].Name
	})

	for i := range s.Relays {
		badges := ""

		switch s.Relays[i].Listener.ProxyType() {
		case localrelay.ProxyTCP:
			badges += "\x1b[90m [TCP] \x1b[0m"
		case localrelay.ProxyHTTP:
			badges += "\x1b[90m [HTTP] \x1b[0m"
		case localrelay.ProxyHTTPS:
			badges += "\x1b[90m [HTTPS] \x1b[0m"
		}

		if s.Relays[i].ProxyEnabled {
			badges += "\x1b[92m [PROXY] \x1b[0m"
		}

		Printf("  \x1b[90m%.2d\x1b[0m: %s %s\r\n      %s -> %s\r\n", i+1, s.Relays[i].Name, badges, s.Relays[i].Listener, s.Relays[i].Destination)
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

func formatDuration(d time.Duration) string {
	if d.Minutes() < 1 {
		return strconv.FormatFloat(d.Seconds(), 'f', 2, 64) + " secs"
	}

	if d.Hours() < 1 {
		return strconv.FormatFloat(d.Minutes(), 'f', 2, 64) + " minutes"
	}

	if d.Hours()/24 < 1 {
		return strconv.FormatFloat(d.Hours(), 'f', 2, 64) + " hours"
	}

	return strconv.FormatFloat(d.Hours()/24, 'f', 2, 64) + " days"
}
