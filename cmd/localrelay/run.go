package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-compile/localrelay"
	"github.com/kardianos/service"
	"github.com/naoina/toml"
	"github.com/pkg/errors"
	"golang.org/x/net/proxy"
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
	configDirSuffix = ".localrelay/"
)

func runRelays(opt *options, i int, cmd []string) error {

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	relayPaths := make([]string, 0, len(cmd[i+1:]))

	// Read all relay config files and decode them
	relays := make([]Relay, 0, len(cmd[i+1:]))
	for _, file := range cmd[i+1:] {

		// if @ used as prefix grab the file from the user profile's
		// config location
		if strings.HasPrefix(file, "@") {
			file = filepath.Join(home, configDirSuffix, file[1:])
		}

		f, err := os.Open(file)
		if err != nil {
			return errors.Wrapf(err, "file:%q", file)
		}

		var relay Relay
		if err := toml.NewDecoder(f).Decode(&relay); err != nil {
			f.Close()
			return err
		}

		relays = append(relays, relay)
		// append path here so we validate the config first before sending to
		// service.
		relayPaths = append(relayPaths, file)

		f.Close()
	}

	if len(relays) == 0 {
		fmt.Println("[WARN] No relay configs provided.")
		return nil
	}

	fmt.Printf("Loaded: %d relays\n", len(relays))

	// if detach is enable fork process and start daemon
	if opt.detach {
		running, err := daemonService.Status()
		if err != nil {
			return err
		}

		if running != service.StatusRunning {
			fmt.Println("[Info] Service not running.")

			if err := daemonService.Start(); err != nil {
				log.Fatalf("[Error] Failed to start service: %s\n", err)
			}

			fmt.Println("[Info] Service has been started.")
		}

		return serviceRun(relayPaths)
	}

	return launchRelays(relays, true)
}

func launchRelays(relays []Relay, wait bool) error {
	// TODO: listen for sigterm signal and softly shutdown

	wg := sync.WaitGroup{}
	activeRelays = make(map[string]*localrelay.Relay, len(relays))
	logDescriptors = make(map[string]*io.Closer, len(relays))

	for i, r := range relays {
		fmt.Printf("[Info] [Relay:%d] Starting %q on %q\n", i+1, r.Name, r.Host)

		if r.Proxy.Host != "" && strings.ToLower(r.Proxy.Protocol) != "socks5" {
			fmt.Printf("[Warn] Proxy type %q not supported.\n", r.Proxy.Protocol)
			return nil
		}

		w := os.Stdout
		if r.Logging != "stdout" {
			fmt.Printf("[Info] [Relay:%s] Log output writing to: %q\n", r.Name, r.Logging)

			f, err := os.OpenFile(r.Logging, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
			if err != nil {
				return err
			}

			addLogDescriptor(f, r.Name)
			w = f
		}

		relay := localrelay.New(r.Name, r.Host, r.Destination, w)

		switch r.Kind {
		case localrelay.ProxyTCP, localrelay.ProxyFailOverTCP:
			// If proxy enabled
			if r.Proxy.Host != "" && strings.ToLower(r.Proxy.Protocol) == "socks5" {

				auth := &proxy.Auth{
					User:     r.Proxy.Username,
					Password: r.Proxy.Password,
				}

				// If auth not set make it nil
				if r.Proxy.Username == "" {
					auth = nil
				}

				prox, err := proxy.SOCKS5("tcp", r.Proxy.Host, auth, nil)
				if err != nil {
					panic(err)
				}

				relay.SetProxy(&prox)
			}

			if r.Kind == localrelay.ProxyFailOverTCP {
				relay.SetFailOverTCP()
				relay.DisableProxy(r.ProxyIgnore...)
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
		case localrelay.ProxyHTTP, localrelay.ProxyHTTPS:
			// Convert the relay from the default: TCP to a HTTP server
			err := relay.SetHTTP(http.Server{
				// Middle ware can be set here
				Handler: localrelay.HandleHTTP(relay),

				ReadTimeout:  time.Second * 60,
				WriteTimeout: time.Second * 60,
				IdleTimeout:  time.Second * 120,
			})

			if err != nil {
				panic(err)
			}

			if relay.ProxyType == localrelay.ProxyHTTPS {
				// Set TLS certificates & make relay HTTPS
				relay.SetTLS(r.Certificate, r.Key)
			}

			// If proxy enabled
			if r.Proxy.Host != "" && strings.ToLower(r.Proxy.Protocol) == "socks5" {

				userinfo := url.UserPassword(r.Proxy.Username, r.Proxy.Password)
				prox, err := url.Parse(r.Proxy.Protocol + "://" + r.Proxy.Host)
				if err != nil {
					panic(err)
				}

				prox.User = userinfo

				relay.SetClient(&http.Client{
					Transport: &http.Transport{
						Proxy: http.ProxyURL(prox),
					},

					Timeout: time.Second * 120,
				})
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

		}
	}

	if wait {
		wg.Wait()
		fmt.Println("[Info] All relays closed.")
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
