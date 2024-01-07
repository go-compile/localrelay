package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/tam7t/hpkp"

	"github.com/go-compile/localrelay/v2"
)

func main() {
	// Create new relay
	// nextcloud is the name of the relay. Note this can be called anything
	// 127.0.0.1:90 is the address the relay will listen on. E.g. you connect via localhost:90
	// http://example.com is the destination address, this can be a remote server
	r, err := localrelay.New("http-relay", os.Stdout, "http://127.0.0.1:90", "https://example.com")
	if err != nil {
		panic(err)
	}

	// Convert the relay from the default: TCP to a HTTP server
	err = r.SetHTTP(&http.Server{
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
			"Xs+pjRp23QkmXeH31KEAjM1aWvxpHT6vYy+q2ltqtaM=",
			"RQeZkB42znUfsDIIFWIRiYEcKl7nHwNFwWCrnMMJbVc=",
		},
	})

	client := &http.Client{}
	dialConf := &hpkp.DialerConfig{
		Storage:   s,
		PinOnly:   true,
		TLSConfig: nil,
		Reporter: func(p *hpkp.PinFailure, reportUri string) {
			fmt.Printf("Certificate did not match locked certificate. Expected: %s got %s\n",
				s.Lookup("example.com").Sha256Pins, returnedCertificatePin(),
			)
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

func returnedCertificatePin() (fingerprints []string) {
	conn, err := tls.Dial("tcp", "example.com:443", &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		panic(err)
	}

	for _, cert := range conn.ConnectionState().PeerCertificates {
		fingerprints = append(fingerprints, hpkp.Fingerprint(cert))
	}

	return fingerprints
}
