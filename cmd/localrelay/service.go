package main

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kardianos/service"
)

func (p daemon) Start(s service.Service) error {
	Println(s.String() + " started")
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

	Println(s.String() + " stopped")
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

	home := configSystemDir()
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

		if !relay.AutoRestart {
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

// configSystemDir returns the parent config dir depending on the system.
// A additional folder will be created within as a child inode.
func configSystemDir() string {
	if runtime.GOOS == "windows" {
		return "C:\\ProgramData"
	}

	return "/etc"
}

func relaysDir() string {
	return filepath.Join(configSystemDir(), configDirSuffix)
}

// securityCheckBinary checks for common issues yet is not comprehensive
func securityCheckBinary() (bool, string, error) {
	exe, err := os.Executable()
	if err != nil {
		return false, "", err
	}

	if runtime.GOOS != "windows" {
		// validate file location
		if !(strings.HasPrefix(exe, "/usr/bin/") ||
			strings.HasPrefix(exe, "/usr/sbin/") ||
			strings.HasPrefix(exe, "/usr/local/bin") ||
			strings.HasPrefix(exe, "/bin/") ||
			strings.HasPrefix(exe, "/sbin/")) {
			return false, "Binary is outside of an appropriate bin directory.", nil
		}
	}

	fileStat, err := os.Stat(exe)
	if err != nil {
		return false, "Could not sta binary.", err
	}

	owner, err := fileOwnership(fileStat)
	if err != nil {
		return false, "Could not attain file ownership information.", err
	}

	// if file owned by root
	if owner != "0" && runtime.GOOS != "windows" {
		return false, "Binary is not solely owned by root/administrator.", err
	}

	// TODO: check if non-owners have write access

	return true, "", nil
}
