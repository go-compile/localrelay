package main

import (
	"os"

	"github.com/go-compile/localrelay"
	"golang.org/x/net/proxy"
)

func main() {
	// Create new relay
	// nextcloud is the name of the relay. Note this can be called anything
	// 127.0.0.1:90 is the address the relay will listen on. E.g. you connect via localhost:90
	// localhost:8080 is the destination address, if this fails to dial it will call:
	// fb3709b05939f04cf2e92f7d0897fc2596f9ad0b8a9ea855c7bfebaae892.onion:80
	r := localrelay.New("nextcloud", "127.0.0.1:90", "localhost:440,fb3709b05939f04cf2e92f7d0897fc2596f9ad0b8a9ea855c7bfebaae892.onion:80", os.Stdout)
	r.SetFailOverTCP()

	// This disables the Tor SOCKS proxy for the remote "localhost:440" but keeps it active
	// for all other remotes such as the onion address.
	r.DisableProxy(0)

	// Create a new SOCKS5 proxy

	// 127.0.0.1:9050 is the Tor SOCKS5 proxy address on all opperating systems
	// other than Windows. On windows it's 9150 however, if you run Tor as a
	// service on Windows (tor.exe not the whole Tor Browser Bundle) the address
	// will be 9050
	prox, err := proxy.SOCKS5("tcp", "127.0.0.1:9150", nil, nil)
	if err != nil {
		panic(err)
	}

	// SetProxy tells the relay you want to use a proxy
	r.SetProxy(&prox)

	// Starts the relay server
	panic(r.ListenServe())
}
