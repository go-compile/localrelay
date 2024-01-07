package main

import (
	"net/url"
	"os"

	"github.com/go-compile/localrelay/v2"
)

func main() {
	// Create new relay
	// onion-service is the name of the relay. Note this can be called anything
	// 127.0.0.1:90 is the address the relay will listen on. E.g. you connect via localhost:90
	// 2gzyxa5ihm7nsggfxnu52rck2vv4rvmdlkiu3zzui5du4xyclen53wid.onion:80 this can be a normal IP
	// address or even a onion if you're using Tor
	r, err := localrelay.New("onion-service", os.Stdout, "tcp://127.0.0.1:90", "tcp://2gzyxa5ihm7nsggfxnu52rck2vv4rvmdlkiu3zzui5du4xyclen53wid.onion:80?proxy=tor")
	if err != nil {
		panic(err)
	}

	// Create a new SOCKS5 proxy

	// 127.0.0.1:9050 is the Tor SOCKS5 proxy address on all opperating systems
	// other than Windows. On windows it's 9150 however, if you run Tor as a
	// service on Windows (tor.exe not the whole Tor Browser Bundle) the address
	// will be 9050
	// Route traffic through Tor
	torProxy, err := url.Parse("socks5://127.0.0.1:9050")
	if err != nil {
		panic(err)
	}

	r.SetProxy(map[string]localrelay.ProxyURL{"tor": {
		URL: torProxy,
	}})

	// Start the relay server
	panic(r.ListenServe())
}
