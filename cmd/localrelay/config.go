package main

import "github.com/go-compile/localrelay"

// Relay is a config for a relay server
type Relay struct {
	Name        string
	Host        string
	Destination string
	// Kind is ProxyType; TCP, HTTP, HTTPS
	Kind localrelay.ProxyType
	// Logging; stdout, ./filename.log
	Logging string

	// Certificate for TLS
	Certificate string
	Key         string

	Proxy *Proxy
}

// Proxy is used for relay forwarding
type Proxy struct {
	Protocol string
	Host     string

	Username string
	Password string
}
