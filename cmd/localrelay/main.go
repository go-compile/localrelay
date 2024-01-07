package main

import (
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/kardianos/service"
	"github.com/pkg/errors"
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
	// set default output, may be changed if ipc stream is enabled
	stdout = os.Stdout

	serviceConfig := &service.Config{
		Name:        serviceName,
		DisplayName: serviceName,
		Description: serviceDescription,
		Arguments:   []string{"start-service-daemon"},
		Option: service.KeyValue{
			"DelayedAutoStart":       true,
			"OnFailure":              "restart",
			"OnFailureDelayDuration": "5s",
		},
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
		Println(err)
		return
	}

	// if process was forked forward stdout
	if len(opt.ipcPipe) != 0 {
		conn, err := forwardIO(opt)
		if err != nil {
			Println(err)
			return
		}

		defer conn.Close()
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
				Println(err)
			}

			return
		case "run":
			if err := runRelays(opt, i, opt.commands); err != nil {
				Println(err)
			}

			return
			// stop will shutdown the daemon service
		case "stop":
			if !privCommand(true) {
				return
			}

			if len(opt.commands) == 1 {
				if err := s.Stop(); err != nil {
					log.Fatalf("[Error] Failed to stop service: %s\n", err)
				}

				Println("Daemon has been shutdown")
				return
			}

			if err := stopRelay(opt.commands[1]); err != nil {
				log.Fatalf("[Error] Failed to stop service: %s\n", err)
			}

			return

		// install will register the daemon service
		case "install":
			secure, msg, err := securityCheckBinary()
			if err != nil {
				log.Fatal(errors.Wrap(err, "checking binary security"))
			}

			if !secure {
				Printf("WARNING!\n Security issues detected. Installation Blocked!\n"+
					"Please follow localrelay's service installation guide to avoid inadvertently"+
					"exposing your system to security vulnerabilities. It is likely your binary has"+
					"insecure permissions.\n\nAudit Results:\n%s\n", msg)
				return
			}

			if !privCommand(true) {
				return
			}

			if err := s.Install(); err != nil {
				log.Fatalf("[Error] Failed to install service: %s\n", err)
			}

			Println("Daemon service has been installed.")

			return
		case "uninstall":
			if !privCommand(true) {
				return
			}

			if err := s.Uninstall(); err != nil {
				log.Fatalf("[Error] Failed to uninstall service: %s\n", err)
			}

			Println("Daemon service has been uninstalled.")

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
			if !privCommand(true) {
				return
			}

			if err := s.Restart(); err != nil {
				log.Fatalf("[Error] Failed to restart service: %s\n", err)
			}

			Println("Daemon has been restarted")
		case "start":
			if !privCommand(true) {
				return
			}

			if err := s.Start(); err != nil {
				log.Fatalf("[Error] Failed to start service: %s\n", err)
			}

			Println("Daemon has been started")
			return
		case "conns", "connections":
			if !privCommand(true) {
				return
			}

			if err := displayOpenConns(opt, false); err != nil {
				Println(err)
				os.Exit(1)
			}

			return
		case "ips":
			if !privCommand(true) {
				return
			}

			if err := displayOpenConns(opt, true); err != nil {
				Println(err)
				os.Exit(1)
			}

			return
		case "status":
			if !privCommand(true) {
				return
			}

			if err := relayStatus(); err != nil {
				Println(err)
			}

			return
			// TODO: add logs. Connect via IPC and show live view
		case "metrics", "monitor":
			if !privCommand(false) {
				return
			}

			if err := relayMetrics(opt); err != nil {
				Println(err)
			}

			return
		case "drop":
			if !privCommand(true) {
				return
			}

			if err := dropConns(opt); err != nil {
				Println(err)
			}
			return
		case "dropip":
			if !privCommand(true) {
				return
			}

			if err := dropConnsIP(opt); err != nil {
				Println(err)
			}
			return
		case "droprelay":
			if !privCommand(true) {
				return
			}

			if err := dropConnsRelay(opt); err != nil {
				Println(err)
			}
			return
		default:
			Printf("Unrecognised command %q\n", opt.commands[i])
			return
		}
	}
}

// privCommand will handle permission elevation.
// If bool is false exit program without errors.
// If bool true the user has permission.
func privCommand(autoFork bool) bool {
	if !runningAsRoot() {
		if runtime.GOOS == "windows" && autoFork {
			// UAC prompt
			if err := fork(); err != nil {
				Println(err)
				return false
			}
		} else {
			Println("Elevated privileges required.")
		}

		return false
	}

	return true
}
