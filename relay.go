package localrelay

import (
	"io"
	"net"
	"net/http"
	"sync"

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

	proxies []*proxy.Dialer

	logger *Logger

	// close is linked to the listener
	close io.Closer

	// Metrics is used to store information such as upload/download
	// and other statistics
	*Metrics

	// http relay section
	server     http.Server
	httpClient *http.Client

	// TLS settings
	certificateFile string
	keyFile         string

	running bool
	m       sync.Mutex
}

const (
	// ProxyTCP is for raw TCP forwarding
	ProxyTCP ProxyType = iota
	// ProxyHTTP creates a HTTP server and forwards the traffic to
	// either a HTTP or HTTPs server
	ProxyHTTP
	// ProxyHTTPS is the same as HTTP but listens on TLS
	ProxyHTTPS

	// VERSION uses semantic versioning
	VERSION = "v0.2.0"
)

var (
	// ErrUnknownProxyType is returned when a relay has a proxy type which is invalid
	ErrUnknownProxyType = errors.New("unknown proxytype used in creation of relay")
	// ErrAddrNotMatch is returned when a server object has a addr which is not nil
	// and does not equal the relay's address
	ErrAddrNotMatch = errors.New("addr does not match the relays host address")
)

// New creates a new TCP relay
func New(name, host, destination string, logger io.Writer) *Relay {

	return &Relay{
		Name:        name,
		Host:        host,
		ForwardAddr: destination,
		ProxyType:   ProxyTCP,

		Metrics: &Metrics{
			// Preallocate array with capacity of 10
			dialTimes: make([]int64, 0, 10),
		},

		httpClient: http.DefaultClient,

		logger: NewLogger(logger, name),
	}
}

// Running returns true if relay is running
func (r *Relay) Running() bool {
	r.m.Lock()
	defer r.m.Unlock()

	return r.running
}

func (r *Relay) setRunning(toggle bool) {
	r.m.Lock()
	defer r.m.Unlock()

	r.running = toggle
}

// SetHTTP is used to set the relay as a type HTTP relay
// addr will auto be set in the server object if left blank
func (r *Relay) SetHTTP(server http.Server) error {
	r.ProxyType = ProxyHTTP

	// Auto set addr if left blank
	if server.Addr == "" {
		server.Addr = r.Host
	} else if server.Addr != r.Host {
		return ErrAddrNotMatch
	}

	r.server = server

	return nil
}

// SetClient will set the http client used by the relay
func (r *Relay) SetClient(client *http.Client) {
	r.httpClient = client
}

// SetTLS sets the TLS certificates for use in the ProxyHTTPS relay.
// This function will upgrade this relay to a HTTPS relay
func (r *Relay) SetTLS(certificateFile, keyFile string) {
	r.certificateFile = certificateFile
	r.keyFile = keyFile

	r.ProxyType = ProxyHTTPS
}

// SetProxy sets the proxy dialer to be used
// proxy.SOCKS5() can be used to setup a socks5 proxy
// or a list of proxies
func (r *Relay) SetProxy(dialer ...*proxy.Dialer) {
	r.proxies = dialer
}

// Close will close the relay's listener
func (r *Relay) Close() error {
	return r.close.Close()
}

// ListenServe will start a listener and handle the incoming requests
func (r *Relay) ListenServe() error {

	defer func() {
		r.logger.Info.Printf("STOPPING: %q on %q\n", r.Name, r.Host)
		r.setRunning(false)
	}()

	r.setRunning(true)

	r.logger.Info.Printf("STARTING: %q on %q\n", r.Name, r.Host)

	l, err := listener(r)
	if err != nil {
		return err
	}

	switch r.ProxyType {
	case ProxyTCP:
		r.close = l

		return relayTCP(r, l)
	case ProxyHTTP:
		r.close = l

		return relayHTTP(r, l)
	case ProxyHTTPS:
		r.close = l

		return relayHTTPS(r, l)
	default:
		l.Close()

		return ErrUnknownProxyType
	}
}

// Serve lets you set your own listener and then serve on it
func (r *Relay) Serve(l net.Listener) error {
	defer func() {
		r.logger.Info.Printf("STOPPING: %q on %q\n", r.Name, r.Host)
		r.setRunning(false)
	}()

	r.setRunning(true)

	r.logger.Info.Printf("STARTING: %q on %q\n", r.Name, r.Host)
	r.close = l

	switch r.ProxyType {
	case ProxyTCP:
		return relayTCP(r, l)
	case ProxyHTTP:
		return relayHTTP(r, l)
	case ProxyHTTPS:
		return relayHTTPS(r, l)
	default:
		return ErrUnknownProxyType
	}
}
