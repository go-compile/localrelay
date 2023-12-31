package main

import "github.com/go-compile/localrelay/v2"

// Relay is a config for a relay server
type Relay struct {
	Name     string
	Listener localrelay.TargetLink
	// DisableAutoStart will stop the daemon from auto starting this relay
	AutoRestart bool
	// Logging; stdout, ./filename.log
	Logging string

	Destinations []localrelay.TargetLink

	Tls     TLS
	Proxies *Proxy
}

// TLS is used when configuring https proxies
type TLS struct {
	Certificate string
	Private     string
}

// Proxy is used for relay forwarding
type Proxy struct {
	Protocol string
	Address  string
	Username string
	Password string
}
