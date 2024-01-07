package main

import (
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/go-compile/localrelay/v2"
)

func main() {
	// Create new relay
	// nextcloud is the name of the relay. Note this can be called anything
	// 127.0.0.1:90 is the address the relay will listen on. E.g. you connect via localhost:90
	// https://check.torproject.org is the destination address, this can be a remote server
	r, err := localrelay.New("http-spoof", os.Stdout, "http://127.0.0.1:90", "https://check.torproject.org?proxy=tor")
	if err != nil {
		panic(err)
	}

	// Route traffic through Tor
	torProxy, err := url.Parse("socks5://127.0.0.1:9050")
	if err != nil {
		panic(err)
	}

	r.SetProxy(map[string]localrelay.ProxyURL{"tor": {
		URL: torProxy,
	}})

	// Convert the relay from the default: TCP to a HTTP server
	err = r.SetHTTP(&http.Server{
		// On each request this middleware will be executed
		// changing the useragent and the accept language
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// Spoof user-agent
			req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_12) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4891.171 Safari/537.36")

			// Spoof accept language to en-US
			req.Header.Set("Accept-Language", "en-US,en;q=0.5")

			// Then send request to localrelay
			localrelay.HandleHTTP(r)(w, req)
		}),

		ReadTimeout:  time.Second * 15,
		WriteTimeout: time.Second * 15,
		IdleTimeout:  time.Second * 30,
	})

	if err != nil {
		panic(err)
	}

	// Starts the relay server
	panic(r.ListenServe())
}
