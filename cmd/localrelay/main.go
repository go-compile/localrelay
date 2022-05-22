package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/go-compile/localrelay"
	"github.com/kardianos/service"
)

const (
	// VERSION uses semantic versioning
	VERSION = localrelay.VERSION
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
		case "stop":
			if err := s.Stop(); err != nil {
				log.Fatalf("[Error] Failed to stop service: %s\n", err)
			}

			fmt.Println("Daemon has been shutdown")
			return
		case "install":
			if err := s.Install(); err != nil {
				fmt.Println("[Warn] Administrator privileges are required to install.")
				log.Fatalf("[Error] Failed to install service: %s\n", err)
			}

			fmt.Println("Daemon service has been installed.")

			return
		case "start-service-daemon":
			if err := s.Run(); err != nil {
				log.Fatalf("[Error] Failed to run service: %s\n", err)
			}

			return
		case "restart":
			if err := relayStatus(); err != nil {
				fmt.Println(err)
				return
			}

			// if err := forkDeamon(); err != nil {
			// 	fmt.Println("Could not restart/fork daemon")
			// 	return
			// }

			fmt.Println("Daemon has been restarted")
			return
		case "status":
			if err := relayStatus(); err != nil {
				fmt.Println(err)
			}

			return
			// TODO: add logs. Connect via IPC and show live view
			// TODO: add metrics with -interval=5s
			// For attached and detached instances
			// TODO: add autoload home dir config argument, use in addition of run configs
		default:
			fmt.Printf("Unrecognised command %q\n", opt.commands[i])
			return
		}
	}
}
