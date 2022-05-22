package main

import (
	"fmt"
	"log"

	"github.com/kardianos/service"
)

func (p daemon) Start(s service.Service) error {
	fmt.Println(s.String() + " started")
	go p.run()
	return nil
}

func (p daemon) Stop(s service.Service) error {
	for _, r := range runningRelays() {
		log.Printf("[Info] Closing relay: %s\n", r.Name)
		if err := r.Close(); err != nil {
			log.Printf("[Error] Closing relay: %s with error: %s\n", r.Name, err)
		}
	}

	log.Printf("[Info] All relays closed:\n")

	closeLogDescriptors()

	ipcListener.Close()

	fmt.Println(s.String() + " stopped")
	return nil
}

func (p daemon) run() {

	// TODO: auto start relays in app config dir

	// listen to commands over IPC
	if err := IPCListen(); err != nil {
		log.Fatal(err)
	}

}
