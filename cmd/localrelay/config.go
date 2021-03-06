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
	// ProxyIgnore is a list of destination indexes where
	// the proxy settings should be ignored.
	ProxyIgnore []int

	// DisableAutoStart will stop the daemon from auto starting this relay
	DisableAutoStart bool

	// ProtocolSwitch allows the outbound protocol to be switched
	// this only works for TCP and UDP.
	// NOTE: If a proxy is enabled protocol switching is disabled
	ProtocolSwitch []string
}

// Proxy is used for relay forwarding
type Proxy struct {
	Protocol string
	Host     string

	Username string
	Password string
}
