package main

import (
	"os"
	"time"

	"github.com/go-compile/localrelay"
)

func main() {
	// Set the remote dial time out globally.
	// Remote dial is only used for remotes when proxies aren't in use
	// or are being ignored.
	localrelay.Timeout = time.Second * 2
	// Create new relay
	// nextcloud is the name of the relay. Note this can be called anything
	// 127.0.0.1:90 is the address the relay will listen on. E.g. you connect via localhost:90
	// localhost:8080 is the destination address, this can be a remote server
	r := localrelay.New("nextcloud", "127.0.0.1:90", "localhost:8080", os.Stdout)

	// Starts the relay server
	panic(r.ListenServe())
}
