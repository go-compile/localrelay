package localrelay

import (
	"io"

	"github.com/pkg/errors"
	"golang.org/x/net/proxy"
)

// ProxyType represents what type of proxy the relay is.
//
// Raw TCP is used for just forwarding the raw connection
// to the remote address.
type ProxyType uint8

// Relay represents a reverse proxy and all of its settings
type Relay struct {
	// Name is a generic name which can be assigned to this relay
	Name string
	// Host is the address to listen on
	Host string

	// ForwardAddr is the destination to send the connection
	ForwardAddr string
	// ProxyType is used to forward or manipulate the connection
	ProxyType ProxyType

	proxy *proxy.Dialer

	logger *Logger
}

const (
	// ProxyTCP is for raw TCP forwarding
	ProxyTCP ProxyType = iota
)

var (
	// ErrUnknownProxyType is returned when a relay has a proxy type which is invalid
	ErrUnknownProxyType = errors.New("unknown proxytype used in creation of relay")
)

// New creates a new TCP relay
func New(name, host, destination string, logger io.Writer) *Relay {

	return &Relay{
		Name:        name,
		Host:        host,
		ForwardAddr: destination,
		ProxyType:   ProxyTCP,

		logger: NewLogger(logger, name),
	}
}

// SetProxy sets the proxy dialer to be used
// proxy.SOCKS5() can be used to setup a socks5 proxy
func (r *Relay) SetProxy(dialer proxy.Dialer) {
	r.proxy = &dialer
}

// ListenServe will start a listener and handle the incoming requests
func (r *Relay) ListenServe() error {

	defer r.logger.Info.Printf("STOPPING: %q on %q\n", r.Name, r.Host)

	r.logger.Info.Printf("STARTING: %q on %q\n", r.Name, r.Host)

	switch r.ProxyType {
	case ProxyTCP:
		return relayTCP(r)
	default:
		return ErrUnknownProxyType
	}
}
