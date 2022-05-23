package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

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

	if err := launchAutoStartRelays(); err != nil {
		log.Fatal(err)
	}

	// listen to commands over IPC
	if err := IPCListen(); err != nil {
		log.Fatal(err)
	}

}

func launchAutoStartRelays() error {
	if err := createConfigDir(); err != nil {
		return err
	}

	home := configDir()
	prefix := filepath.Join(home, configDirSuffix)

	// read config dir in home folder
	dir, err := os.ReadDir(prefix)
	if err != nil {
		return err
	}

	relays := make([]Relay, 0, len(dir))

	for _, entry := range dir {
		// ignore all none toml files
		if filepath.Ext(entry.Name()) != ".toml" || entry.IsDir() {
			continue
		}

		file := filepath.Join(home, configDirSuffix, entry.Name())

		relay, err := readRelayConfig(file)
		if err != nil {
			return err
		}

		relays = append(relays, *relay)

		log.Printf("[Launching relay] %q\n", relay.Name)
	}

	if len(relays) == 0 {
		return nil
	}

	return launchRelays(relays, false)
}

func configDir() string {
	if runtime.GOOS == "windows" {
		return "C:\\ProgramData"
	}

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Error fetching home dir:", err)
	}

	return home
}
