package localrelay

import "github.com/go-compile/localrelay/v2"

type Status struct {
	Relays  []localrelay.Relay
	Pid     int
	Version string
	// Metrics contains relay name as the index
	Metrics map[string]Metrics
	// Started is a unix timestamp of when the daemon was created
	Started int64
}

type Metrics struct {
	In, Out, Active, DialAvg  int
	TotalConns, TotalRequests uint64
}

type Connection struct {
	LocalAddr  string
	RemoteAddr string
	Network    string

	RelayName string
	RelayHost string

	ForwardedAddr string

	// Opened is a unix timestamp
	Opened int64
}
