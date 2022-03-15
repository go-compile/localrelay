package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/tam7t/hpkp"

	"github.com/go-compile/localrelay"
)

func main() {
	// Create new relay
	// nextcloud is the name of the relay. Note this can be called anything
	// 127.0.0.1:90 is the address the relay will listen on. E.g. you connect via localhost:90
	// http://example.com is the destination address, this can be a remote server
	r := localrelay.New("https-relay", "127.0.0.1:90", "https://example.com", os.Stdout)

	// Convert the relay from the default: TCP to a HTTP server
	err := r.SetHTTP(http.Server{
		// Middle ware can be set here
		Handler: localrelay.HandleHTTP(r),

		ReadTimeout:  time.Second * 15,
		WriteTimeout: time.Second * 15,
		IdleTimeout:  time.Second * 30,
	})

	if err != nil {
		panic(err)
	}

	// Certificate pinning via https://github.com/tam7t/hpkp
	s := hpkp.NewMemStorage()
	s.Add("example.com", &hpkp.Header{
		Permanent: true,
		Sha256Pins: []string{
			"WoiWRyIOVNa9ihaBciRSC7XHjliYS9VwUGOIud4PB18=",
		},
	})
	client := &http.Client{}
	dialConf := &hpkp.DialerConfig{
		Storage:   s,
		PinOnly:   true,
		TLSConfig: nil,
		Reporter: func(p *hpkp.PinFailure, reportUri string) {
			fmt.Printf("Certificate did not match locked certificate. Expected: %s\n", s.Lookup("example.com").Sha256Pins)
		},
	}

	client.Transport = &http.Transport{
		DialTLS: dialConf.NewDialer(),
	}

	// Set the http client for the relay
	r.SetClient(client)

	// Starts the relay server
	panic(r.ListenServe())
}
