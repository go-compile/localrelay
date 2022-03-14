package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-compile/localrelay"
	"github.com/naoina/toml"
	"github.com/pkg/errors"
	"golang.org/x/net/proxy"
)

func runRelays(opt *options, i int, cmd []string) error {

	// Read all relay config files and decode them
	relays := make([]Relay, 0, len(cmd[i+1:]))
	for _, file := range cmd[i+1:] {

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

		f.Close()
	}

	if len(relays) == 0 {
		fmt.Println("[WARN] No relay configs provided.")
		return nil
	}

	fmt.Printf("Loaded: %d relays\n", len(relays))

	return launchRelays(relays)
}

func launchRelays(relays []Relay) error {

	wg := sync.WaitGroup{}

	for i, r := range relays {
		fmt.Printf("[Info] [Relay:%d] Starting %q on %q\n", i+1, r.Name, r.Host)

		// TODO: add logging to file
		w := os.Stdout
		relay := localrelay.New(r.Name, r.Host, r.Destination, w)

		switch r.Kind {
		case localrelay.ProxyTCP:
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

			wg.Add(1)
			go func(relay *localrelay.Relay) {
				if err := relay.ListenServe(); err != nil {
					log.Println("[Error] ", err)
				}

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

			wg.Add(1)
			go func(relay *localrelay.Relay) {
				if err := relay.ListenServe(); err != nil {
					log.Println("[Error] ", err)
				}

				wg.Done()
			}(relay)

		}
	}

	wg.Wait()
	fmt.Println("[Info] All relays closed.")
	return nil
}
