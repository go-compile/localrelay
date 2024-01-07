package main

import (
	"fmt"
	"os"
	"time"

	"github.com/go-compile/localrelay/v2"
)

func main() {
	// Create new relay
	r, err := localrelay.New("nextcloud", os.Stdout, "tcp://127.0.0.1:90", "tcp://localhost:8080")
	if err != nil {
		panic(err)
	}

	// Close relay after 15 seconds
	go func() {
		time.Sleep(time.Second * 15)
		r.Close()
	}()

	// Start the relay and handle requests
	if err := r.ListenServe(); err != nil {
		panic(err)
	}

	fmt.Println("Server closed")
}
