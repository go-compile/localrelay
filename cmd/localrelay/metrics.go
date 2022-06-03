package main

import (
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"
	"time"

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

	relays := []string{}

	// build filter list for relays
	if len(opt.commands) > 1 {
		for _, relayName := range opt.commands[1:] {
			if _, ok := status.Metrics[strings.ToLower(relayName)]; !ok {
				fmt.Printf("Relay %q is not running.\n", relayName)
				return nil
			}

			relays = append(relays, relayName)
		}
	}

	// lock := sync.Mutex{}

	sig := make(chan os.Signal, 1)
	go func() {
		signal.Notify(sig, os.Interrupt)
	}()

	for {
		select {
		case <-sig:
			// make a guess how far to move the cursor
			if len(relays) != 0 {
				fmt.Printf("\x1b[%dB", len(relays)+2)
			} else {
				fmt.Printf("\x1b[%dB", len(status.Metrics)+2)
			}

			return nil
		default:
			status, err := serviceStatus()
			if err != nil {
				return err
			}

			metrics := make([]namedMetrics, 0, len(status.Metrics))
			for k, m := range status.Metrics {
				metrics = append(metrics, namedMetrics{k, m})
			}

			// sort by bandwidth
			sort.SliceStable(metrics, func(i, j int) bool {
				return metrics[i].name > metrics[j].name
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

			fmt.Printf("\r\n\x1b[2K [Running Relays: %d]\r\n", len(status.Metrics))
			fmt.Printf("\x1b[%dA", count+2)
			time.Sleep(opt.interval)
		}
	}
}

func printMetrics(name string, m metrics) {
	fmt.Printf("\x1b[2K%s: [In/Out:%s/%s] [DialAvg:%dms] [Active:%d] [Total:%d]\r\n", name, formatBytes(m.In), formatBytes(m.Out), m.DialAvg, m.Active, m.TotalConns+m.TotalRequests)
}
