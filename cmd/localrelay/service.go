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

	// TODO: listen to signals for reload from systemctl

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

		if relay.DisableAutoStart {
			log.Printf("[Ignoring Relay] %q\n", relay.Name)
			continue
		}

		log.Printf("[Launching relay] %q\n", relay.Name)
	}

	if len(relays) == 0 {
		return nil
	}

	return launchRelays(relays, false)
}

// configDir returns the parent config dir depending on the system.
// A additional folder will be created within as a child inode.
func configDir() string {
	if runtime.GOOS == "windows" {
		return "C:\\ProgramData"
	}

	return "/etc"
}
