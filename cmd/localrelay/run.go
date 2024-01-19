package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-compile/localrelay/v2"
	"github.com/kardianos/service"
	"github.com/naoina/toml"
	"github.com/pkg/errors"
)

var (
	// activeRelays is a list of relays being ran
	activeRelays  map[string]*localrelay.Relay
	activeRelaysM sync.Mutex

	// logDescriptors is a list of relay name to file descriptor
	// this is used when shutting down.
	logDescriptors map[string]*io.Closer

	forkIdentifier = "exec.signal-forked-process-true"

	// configDirSuffix is prepended with the user's home dir.
	// This is where the relay configs are stored.
	configDirSuffix = "localrelay/"

	ErrInvalidRelayName = errors.New("invalid relay name")
)

func init() {
	activeRelays = make(map[string]*localrelay.Relay, 3)
	logDescriptors = make(map[string]*io.Closer, 3)
}

func runRelays(opt *options, i int, cmd []string) error {
	home := configSystemDir()

	relayPaths := make([]string, 0, len(cmd[i+1:]))

	// Read all relay config files and decode them
	relays := make([]Relay, 0, len(cmd[i+1:]))
	for _, file := range cmd[i+1:] {

		// if @ used as prefix grab the file from the user profile's
		// config location
		if strings.HasPrefix(file, "@") {
			file = filepath.Join(home, configDirSuffix, file[1:])
		}

		file, err := filepath.Abs(file)
		if err != nil {
			return err
		}

		relay, err := readRelayConfig(file)
		if err != nil {
			return err
		}

		relays = append(relays, *relay)
		// append path here so we validate the config first before sending to
		// service.
		relayPaths = append(relayPaths, file)

	}

	if len(relays) == 0 {
		Println("[WARN] No relay configs provided.")
		return nil
	}

	Printf("Loaded: %d relays\n", len(relays))

	// if detach is enable fork process and start daemon
	if opt.detach {
		running, err := daemonService.Status()
		if err != nil {
			return errors.Wrap(err, "fetching service status")
		}

		if running != service.StatusRunning {
			Println("[Info] Service not running.")

			if err := daemonService.Start(); err != nil {
				log.Fatalf("[Error] Failed to start service: %s\n", err)
			}

			Println("[Info] Service has been started.")

			// wait for process to launch
			time.Sleep(time.Millisecond * 50)
		}

		return serviceRun(relayPaths)
	}

	return launchRelays(relays, true)
}

func launchRelays(relays []Relay, wait bool) error {
	// TODO: listen for sigterm signal and softly shutdown

	wg := sync.WaitGroup{}

	for i := range relays {
		r := relays[i]

		if !validateName(r.Name) {
			return ErrInvalidRelayName
		}

		Printf("[Info] [Relay:%d] Starting %q on %q\n", i+1, r.Name, r.Listener)

		w := io.MultiWriter(newLogger(r.Name), os.Stdout)
		if !isService {
			// not running as a serivce
			w = os.Stdout
		}

		// was a custom file provided?
		if r.Logging != "stdout" && r.Logging != "default" && r.Logging != "" {
			if isService {
				Printf("[Info] [Relay:%s] Custom log files are not permitted when running in daemon mode\n", r.Name)
			} else {
				Printf("[Info] [Relay:%s] Log output writing to: %q\n", r.Name, r.Logging)

				f, err := os.OpenFile(r.Logging, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
				if err != nil {
					return err
				}

				addLogDescriptor(f, r.Name)
				w = io.MultiWriter(f, os.Stdout)
			}
		}

		relay, err := localrelay.New(r.Name, w, r.Listener, r.Destinations...)
		if err != nil {
			return err
		}

		// ===== set proxies
		proxMap := make(map[string]localrelay.ProxyURL)
		for proxyName, proxyConf := range r.Proxies {
			if strings.ToLower(proxyConf.Protocol) != "socks5" {
				return errors.New("Socks5 is the only supported proxy type")
			}

			proxyURL, err := url.Parse(proxyConf.Protocol + "://" + proxyConf.Address)
			if err != nil {
				return err
			}

			if len(proxyConf.Username) > 0 || len(proxyConf.Password) > 0 {
				proxyURL.User = url.UserPassword(proxyConf.Username, proxyConf.Password)
			}

			proxMap[proxyName] = localrelay.NewProxyURL(proxyURL)
		}

		if len(proxMap) > 0 {
			relay.SetProxy(proxMap)
		}

		if r.Loadbalance.Enabled {
			relay.SetLoadbalance(true)
		}

		switch r.Listener.ProxyType() {
		case localrelay.ProxyTCP, localrelay.ProxyUDP:
			addRelay(relay)
			wg.Add(1)
			go func(relay *localrelay.Relay) {
				if err := relay.ListenServe(); err != nil {
					log.Println("[Error] ", err)
				}

				removeRelay(relay.Name)
				removeLogDescriptor(r.Name)
				wg.Done()
			}(relay)
		case localrelay.ProxyHTTP, localrelay.ProxyHTTPS:
			// Convert the relay from the default: TCP to a HTTP server
			err := relay.SetHTTP(&http.Server{
				// Middle ware can be set here
				Handler: localrelay.HandleHTTP(relay),

				ReadTimeout:  time.Second * 60,
				WriteTimeout: time.Second * 60,
				IdleTimeout:  time.Second * 120,
			})

			if err != nil {
				panic(err)
			}

			if relay.Listener.ProxyType() == localrelay.ProxyHTTPS {
				// Set TLS certificates & make relay HTTPS
				relay.SetTLS(r.Tls.Certificate, r.Tls.Private)
			}

			addRelay(relay)
			wg.Add(1)
			go func(relay *localrelay.Relay) {
				if err := relay.ListenServe(); err != nil {
					log.Println("[Error] ", err)
				}

				removeRelay(relay.Name)
				removeLogDescriptor(r.Name)
				wg.Done()
			}(relay)
		default:
			return errors.New("unknown listener type")
		}
	}

	if wait {
		wg.Wait()
		Println("[Info] All relays closed.")
	}

	return nil
}

func addRelay(r *localrelay.Relay) {
	activeRelaysM.Lock()
	activeRelays[r.Name] = r
	activeRelaysM.Unlock()
}

func removeRelay(name string) {
	activeRelaysM.Lock()
	delete(activeRelays, name)
	activeRelaysM.Unlock()
}

func addLogDescriptor(c io.Closer, name string) {
	activeRelaysM.Lock()
	logDescriptors[name] = &c
	activeRelaysM.Unlock()
}

func removeLogDescriptor(name string) {
	activeRelaysM.Lock()
	delete(logDescriptors, name)
	activeRelaysM.Unlock()
}

func closeLogDescriptors() {
	activeRelaysM.Lock()
	for _, c := range logDescriptors {
		closer := *c

		closer.Close()
	}
	activeRelaysM.Unlock()
}

func isRunning(relay string) bool {
	activeRelaysM.Lock()
	defer activeRelaysM.Unlock()

	_, found := activeRelays[relay]
	return found
}

func runningRelays() []*localrelay.Relay {
	activeRelaysM.Lock()

	relays := make([]*localrelay.Relay, 0, len(activeRelays))
	for _, r := range activeRelays {
		relays = append(relays, r)
	}
	activeRelaysM.Unlock()

	return relays
}

func closeDescriptor(path string) error {
	activeRelaysM.Lock()
	defer activeRelaysM.Unlock()

	closer := *logDescriptors[path]
	return closer.Close()
}

// runningRelaysCopy makes a copy instead of returning the
// pointers
func runningRelaysCopy() []localrelay.Relay {
	activeRelaysM.Lock()

	relays := make([]localrelay.Relay, 0, len(activeRelays))
	for _, r := range activeRelays {
		relays = append(relays, *r)
	}
	activeRelaysM.Unlock()

	return relays
}

func readRelayConfig(file string) (*Relay, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, errors.Wrapf(err, "file:%q", file)
	}

	defer f.Close()

	var relay Relay
	if err := toml.NewDecoder(f).Decode(&relay); err != nil {
		return nil, errors.Wrapf(err, "file:%q", file)
	}

	return &relay, nil
}
