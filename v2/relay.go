package localrelay

import (
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/net/proxy"
)

// ProxyType represents what type of proxy the relay is.
//
// Raw TCP is used for just forwarding the raw connection
// to the remote address.
type ProxyType string

// Relay represents a reverse proxy and all of its settings
type Relay struct {
	// Name is a generic name which can be assigned to this relay
	Name string
	// Listener is the address and protocol to listen on.
	// Example A (listen on loopback):
	// tcp://127.0.0.1:443
	// Example B (listen on all interfaces):
	// tcp://0.0.0.0:443
	Listener TargetLink

	// Destination is an array of connection URLs.
	// Example A:
	// tcp://127.0.0.1:443
	// Example B:
	// udp://127.0.0.1:23
	// Example C:
	// tcp://example.com:443?proxy=tor
	Destination []TargetLink

	// ProxyEnabled is set to true when a proxy has been set for this relay
	ProxyEnabled bool
	proxies      map[string]proxy.Dialer
	// remoteProxyIgnore is a list of indexes in the ForwardAddr array
	// to which the proxy settings should be ignored
	remoteProxyIgnore []int

	logger *Logger

	// close is linked to the listener
	close io.Closer

	// Metrics is used to store information such as upload/download
	// and other statistics
	*Metrics

	// http relay section
	httpServer *http.Server
	httpClient *http.Client

	// TLS settings
	certificateFile string
	keyFile         string

	running bool
	m       sync.Mutex

	protocolSwitching map[int]string

	// connPool contains a list of ACTIVE connections
	connPool []*PooledConn
}

// PooledConn allows meta data to be attached to a connection
type PooledConn struct {
	Conn       net.Conn
	RemoteAddr string
	Opened     time.Time
}

const (
	// ProxyTCP is for raw TCP forwarding
	ProxyTCP ProxyType = "tcp"
	// ProxyUDP forwards UDP traffic
	ProxyUDP ProxyType = "udp"
	// ProxyHTTP creates a HTTP server and forwards the traffic to
	// either a HTTP or HTTPs server
	ProxyHTTP ProxyType = "http"
	// ProxyHTTPS is the same as HTTP but listens on TLS
	ProxyHTTPS ProxyType = "https"

	// VERSION uses semantic versioning
	// this version number is for the library not the CLI
	VERSION = "v2.0.0"
)

var (
	// ErrUnknownProxyType is returned when a relay has a proxy type which is invalid
	ErrUnknownProxyType = errors.New("unknown proxytype used in creation of relay")
	// ErrAddrNotMatch is returned when a server object has a addr which is not nil
	// and does not equal the relay's address
	ErrAddrNotMatch = errors.New("addr does not match the relays host address")
)

// New creates a new TCP relay
func New(name string, logger io.Writer, listener TargetLink, destination ...TargetLink) *Relay {

	return &Relay{
		Name:        name,
		Listener:    listener,
		Destination: destination,

		Metrics: &Metrics{
			// Preallocate array with capacity of 10
			dialTimes: make([]int64, 0, 10),
		},

		httpClient: http.DefaultClient,

		logger: NewLogger(logger, name),

		// TODO: remove if redundent
		// protocolSwitching: make(map[int]string, strings.Count(destination, ",")),
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

// ignoreProxySettings returns true if the proxy should be disabled
// for this remote index
func (r *Relay) ignoreProxySettings(remoteIndex int) bool {
	for _, v := range r.remoteProxyIgnore {
		if v == remoteIndex {
			return true
		}
	}

	return false
}

// SetHTTP is used to set the relay as a type HTTP relay
// addr will auto be set in the server object if left blank
func (r *Relay) SetHTTP(server *http.Server) error {
	return errors.New("not implemented") //TODO: remove once setHTTP fixed

	// Auto set addr if left blank
	if server.Addr == "" {
		server.Addr = r.Listener.Addr()
	} else if server.Addr != r.Listener.Addr() {
		return ErrAddrNotMatch
	}

	// if there is a trailing slash strip it
	// if len(r.ForwardAddr) > 1 && r.ForwardAddr[len(r.ForwardAddr)-1] == '/' {
	// 	r.ForwardAddr = r.ForwardAddr[:len(r.ForwardAddr)-1]
	// }

	// FIXME: ^ how does this interact with the new destination refactor?

	r.httpServer = server

	return nil
}

// SetClient will set the http client used by the relay
func (r *Relay) SetClient(client *http.Client) {
	r.httpClient = client

	if r.httpClient.Transport != nil {
		r.ProxyEnabled = true
	}
}

// SetTLS sets the TLS certificates for use in the ProxyHTTPS relay.
// This function will upgrade this relay to a HTTPS relay
func (r *Relay) SetTLS(certificateFile, keyFile string) {
	r.certificateFile = certificateFile
	r.keyFile = keyFile
}

// SetProxy sets the proxy dialer to be used
// proxy.SOCKS5() can be used to setup a socks5 proxy
// or a list of proxies
func (r *Relay) SetProxy(proxies map[string]proxy.Dialer) {
	r.proxies = proxies
	r.ProxyEnabled = true
}

// Close will close the relay's listener
func (r *Relay) Close() error {
	return r.close.Close()
}

// ListenServe will start a listener and handle the incoming requests
func (r *Relay) ListenServe() error {

	defer func() {
		r.logger.Info.Printf("STOPPING: %q on %q\n", r.Name, r.Listener)
		r.setRunning(false)
	}()

	r.setRunning(true)

	r.logger.Info.Printf("STARTING: %q on %q\n", r.Name, r.Listener)

	l, err := listener(r)
	if err != nil {
		return err
	}

	switch r.Listener.ProxyType() {
	case ProxyTCP:
		r.close = l

		return relayTCP(r, l)
	case ProxyUDP:
		r.close = l

		return relayUDP(r, l)
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
		r.logger.Info.Printf("STOPPING: %q on %q\n", r.Name, r.Listener)
		r.setRunning(false)
	}()

	r.setRunning(true)

	r.logger.Info.Printf("STARTING: %q on %q\n", r.Name, r.Listener)
	r.close = l

	switch r.Listener.ProxyType() {
	case ProxyTCP:
		return relayTCP(r, l)
	case ProxyUDP:
		return relayUDP(r, l)
	case ProxyHTTP:
		return relayHTTP(r, l)
	case ProxyHTTPS:
		return relayHTTPS(r, l)
	default:
		return ErrUnknownProxyType
	}
}

// storeConn places the provided net.Conn into the connPoll.
// To remove this conn from the pool, provide it to popConn()
func (r *Relay) storeConn(conn net.Conn) {
	r.m.Lock()
	defer r.m.Unlock()

	r.connPool = append(r.connPool, &PooledConn{conn, "", time.Now()})
}

// popConn removes the provided connection from the conn pool
func (r *Relay) popConn(conn net.Conn) {
	r.m.Lock()
	defer r.m.Unlock()

	for i := 0; i < len(r.connPool); i++ {
		if r.connPool[i].Conn == conn {
			// remove conn
			r.connPool = append(r.connPool[:i], r.connPool[i+1:]...)
			return
		}
	}
}

// setConnRemote will update the conn pool with the remote
func (r *Relay) setConnRemote(conn net.Conn, remote net.Addr) {
	r.m.Lock()
	defer r.m.Unlock()

	for i := 0; i < len(r.connPool); i++ {
		if r.connPool[i].Conn == conn {
			// remove conn
			r.connPool[i].RemoteAddr = remote.String()
			return
		}
	}
}

// GetConns returns all the active connections to this relay
func (r *Relay) GetConns() []*PooledConn {
	r.m.Lock()
	defer r.m.Unlock()

	return r.connPool
}