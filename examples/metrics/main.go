package main

import (
	"fmt"
	"os"
	"time"

	"github.com/go-compile/localrelay"
	"golang.org/x/net/proxy"
)

func main() {
	// Create new relay
	// onion-service is the name of the relay. Note this can be called anything
	// 127.0.0.1:90 is the address the relay will listen on. E.g. you connect via localhost:90
	// 2gzyxa5ihm7nsggfxnu52rck2vv4rvmdlkiu3zzui5du4xyclen53wid.onion:80 this can be a normal IP
	// address or even a onion if you're using Tor
	r := localrelay.New("onion-service", "127.0.0.1:90", "2gzyxa5ihm7nsggfxnu52rck2vv4rvmdlkiu3zzui5du4xyclen53wid.onion:80", os.Stdout)

	// Create a new SOCKS5 proxy

	// 127.0.0.1:9050 is the Tor SOCKS5 proxy address on all opperating systems
	// other than Windows. On windows it's 9150 however, if you run Tor as a
	// service on Windows (tor.exe not the whole Tor Browser Bundle) the address
	// will be 9050
	prox, err := proxy.SOCKS5("tcp", "127.0.0.1:9050", nil, nil)
	if err != nil {
		panic(err)
	}

	// SetProxy tells the relay you want to use a proxy
	r.SetProxy(&prox)

	// Prints metrics every 5 seconds
	go func() {
		for {
			time.Sleep(time.Second * 5)

			active, total := r.Metrics.Connections()
			fmt.Printf("[In/Out: %d/%d] [Active: %d] [Total: %d] [Dialer Avg: %dms]\n", r.Metrics.Download(), r.Metrics.Upload(), active, total, r.Metrics.DialerAvg())
		}
	}()

	// Start the relay server
	panic(r.ListenServe())
}
