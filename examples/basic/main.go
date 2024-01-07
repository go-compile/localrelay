package main

import (
	"os"

	"github.com/go-compile/localrelay/v2"
)

func main() {
	// Create new relay
	// nextcloud is the name of the relay. Note this can be called anything
	// 127.0.0.1:90 is the address the relay will listen on. E.g. you connect via localhost:90
	// localhost:8080 is the destination address, this can be a remote server
	r, err := localrelay.New("nextcloud", os.Stdout, "tcp://127.0.0.1:90", "tcp://localhost:8080")
	if err != nil {
		panic(err)
	}

	// Starts the relay server
	panic(r.ListenServe())
}
