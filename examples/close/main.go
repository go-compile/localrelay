package main

import (
	"fmt"
	"os"
	"time"

	"github.com/go-compile/localrelay"
)

func main() {
	// Create new relay
	r := localrelay.New("nextcloud", "127.0.0.1:90", "localhost:8080", os.Stdout)

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
