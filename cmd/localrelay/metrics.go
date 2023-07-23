package main

import (
	"bytes"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/containerd/console"
	"github.com/pkg/errors"
)

var (
	// ErrRelayNotRunning is returned if the selected relay isn't running
	ErrRelayNotRunning = errors.New("relay not running")
)

type namedMetrics struct {
	name string
	metrics
}

func relayMetrics(opt *options) error {

	status, err := serviceStatus()
	if err != nil {
		return err
	}

	// setting terminal to raw on linux results in SIGTERM not registering
	if runtime.GOOS == "windows" {
		// make terminal raw to allow the use of colour on windows terminals
		current := console.Current()
		defer current.Reset()

		if err := current.SetRaw(); err != nil {
			log.Fatal(err)
		}
	}

	relays := []string{}

	// build filter list for relays
	if len(opt.commands) > 1 {
		for _, relayName := range opt.commands[1:] {
			if !validateName(relayName) {
				Printf("[WARN] Invalid relay name.")
				return nil
			}

			if _, ok := status.Metrics[strings.ToLower(relayName)]; !ok {
				Printf("Relay %q is not running.\n", relayName)
				return nil
			}

			relays = append(relays, relayName)
		}
	}

	sig := make(chan os.Signal, 1)
	go func() {
		signal.Notify(sig, os.Interrupt)

		if runtime.GOOS == "windows" {
			// listen for interrupt
			for {
				buf := make([]byte, 4)
				n, err := os.Stdin.Read(buf)
				if err != nil {
					os.Exit(0)
					return
				}

				// break on keys "q", "ESC", "CTRL + C" etc
				if bytes.Equal(buf[:n], []byte{3}) || bytes.Equal(buf[:n], []byte{8}) ||
					bytes.Equal(buf[:n], []byte{113}) || bytes.Equal(buf[:n], []byte{27}) {
					close(sig)
					break
				}
			}
		}
	}()

	running := 0
	// set ticker to micro second to triger metrics to render instantly
	// then change tricker within case statement to correct interval
	ticker := time.NewTicker(time.Microsecond)
	defer ticker.Stop()

	for {
		select {
		case <-sig:
			// make a guess how far to move the cursor
			if len(relays) != 0 {
				Printf("\x1b[%dB", (len(relays)*2)+2)
			} else {
				Printf("\x1b[%dB", (len(status.Metrics)*2)+2)
			}

			return nil
		case <-ticker.C:
			ticker.Reset(opt.interval)
			status, err := serviceStatus()
			if err != nil {
				return err
			}

			// if relay has gone offline clear bottom of screen
			if x := len(status.Metrics); x < running {
				Printf("\x1b[0J")
			}

			running = len(status.Metrics)
			totalInOut := [2]int{}

			metrics := make([]namedMetrics, 0, running)
			for k, m := range status.Metrics {
				metrics = append(metrics, namedMetrics{k, m})
				totalInOut[0] += m.In
				totalInOut[1] += m.Out
			}

			// sort alphabetically
			sort.SliceStable(metrics, func(i, j int) bool {
				return metrics[i].name < metrics[j].name
			})

			count := len(relays)

			// if not filter present, show all
			if len(relays) == 0 {
				count = len(metrics)

				for _, m := range metrics {
					printMetrics(m.name, m.metrics)
				}
			} else {
				// sort will be based on order of input args
				for _, v := range relays {
					m, ok := status.Metrics[v]
					if !ok {
						return errors.Wrapf(ErrRelayNotRunning, "%s", v)
					}

					printMetrics(v, m)
				}
			}

			Printf("\r\n\x1b[2K  [Running Relays: %d] [In/Out: %s/%s]\r\n", running, formatBytes(totalInOut[0]), formatBytes(totalInOut[1]))
			Printf("\x1b[%dA", (count*2)+2)
		}
	}
}

func printMetrics(name string, m metrics) {
	Printf("\x1b[2K \x1b[90m%s\x1b[0m\r\n\x1b[2K  [In/Out:%s/%s] [DialAvg:%dms] [Active:%d] [Total:%d]\r\n", name, formatBytes(m.In), formatBytes(m.Out), m.DialAvg, m.Active, m.TotalConns+m.TotalRequests)
}
