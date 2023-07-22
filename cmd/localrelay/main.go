package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/kardianos/service"
)

var (
	// VERSION uses semantic versioning
	VERSION = "(Unknown Version)"
	COMMIT  = "0000000"
	BRANCH  = "(Unknown Branch)"
)

var (
	daemonService service.Service
)

func main() {
	serviceConfig := &service.Config{
		Name:        serviceName,
		DisplayName: serviceName,
		Description: serviceDescription,
		Arguments:   []string{"start-service-daemon"},
	}

	prg := &daemon{}
	s, err := service.New(prg, serviceConfig)
	if err != nil {
		log.Fatalf("[Error] Failed to create service: %s\n", err)
	}

	daemonService = s

	opt, err := parseArgs()
	if err == nil && opt == nil {
		return
	}

	if err != nil {
		fmt.Println(err)
		return
	}

	if len(opt.commands) == 0 {
		help()
		return
	}

	for i := 0; i < len(opt.commands); i++ {

		switch strings.ToLower(opt.commands[i]) {
		case "help", "h", "?":
			help()
			return
		case "version", "v":
			version()
			return
		case "new":
			if err := newRelay(opt, i, opt.commands); err != nil {
				fmt.Println(err)
			}

			return
		case "run":
			if err := runRelays(opt, i, opt.commands); err != nil {
				fmt.Println(err)
			}

			return
			// stop will shutdown the daemon service
		case "stop":

			if len(opt.commands) == 1 {
				if err := s.Stop(); err != nil {
					log.Fatalf("[Error] Failed to stop service: %s\n", err)
				}

				fmt.Println("Daemon has been shutdown")
				return
			}

			if err := stopRelay(opt.commands[1]); err != nil {
				log.Fatalf("[Error] Failed to stop service: %s\n", err)
			}

			return

		// install will register the daemon service
		case "install":
			if err := s.Install(); err != nil {
				fmt.Println("[Warn] Administrator privileges are required to install.")
				log.Fatalf("[Error] Failed to install service: %s\n", err)
			}

			fmt.Println("Daemon service has been installed.")

			return
		case "uninstall":
			if err := s.Uninstall(); err != nil {
				fmt.Println("[Warn] Administrator privileges are required to uninstall.")
				log.Fatalf("[Error] Failed to uninstall service: %s\n", err)
			}

			fmt.Println("Daemon service has been uninstalled.")

			return
			// start-service-daemon will run as the service daemon
		case "start-service-daemon":
			log.Printf("[Version] %s (%s.%s)\n", VERSION, BRANCH, COMMIT)
			daemonStarted = time.Now()
			if err := s.Run(); err != nil {
				log.Fatalf("[Error] Failed to run service: %s\n", err)
			}

			return
			// restart will rerun the service but will not restore previously ran relays
		case "restart":
			if err := s.Restart(); err != nil {
				log.Fatalf("[Error] Failed to restart service: %s\n", err)
			}

			fmt.Println("Daemon has been restarted")
		case "start":
			if err := s.Start(); err != nil {
				log.Fatalf("[Error] Failed to start service: %s\n", err)
			}

			fmt.Println("Daemon has been started")
			return
		case "conns", "connections":
			if err := displayOpenConns(opt, false); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			return
		case "ips":
			if err := displayOpenConns(opt, true); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			return
		case "status":
			if err := relayStatus(); err != nil {
				fmt.Println(err)
			}

			return
			// TODO: add logs. Connect via IPC and show live view
		case "metrics", "monitor":
			if err := relayMetrics(opt); err != nil {
				fmt.Println(err)
			}

			return
		case "drop":
			if err := dropAll(); err != nil {
				fmt.Println(err)
			}
			return
		default:
			fmt.Printf("Unrecognised command %q\n", opt.commands[i])
			return
		}
	}
}
