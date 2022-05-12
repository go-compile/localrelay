package localrelay

import (
	"errors"
	"net"
	"strings"
	"time"
)

var (
	// ErrFailConnect will be returned if the remote failed to dial
	ErrFailConnect = errors.New("failed to dial remote")
	// Timeout is only used when dialling without a proxy
	Timeout = time.Second * 5
)

func relayFailOverTCP(r *Relay, l net.Listener) error {

	r.logger.Info.Println("STARTING FAIL OVER TCP RELAY")

	for {
		conn, err := l.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				r.logger.Warning.Println("LISTENER CLOSED")
				return nil
			}

			r.logger.Warning.Println("ACCEPT FAILED: ", err)
			continue
		}

		go handleFailOverTCP(r, conn)
	}
}

func handleFailOverTCP(r *Relay, conn net.Conn) {
	defer func() {
		conn.Close()
		r.Metrics.connections(-1)
	}()

	r.Metrics.connections(1)

	r.logger.Info.Printf("NEW CONNECTION %q ON %q\n", conn.RemoteAddr(), conn.LocalAddr())

	start := time.Now()

	// If using a proxy dial with proxy
	if r.proxies != nil {
		r.logger.Info.Println("CREATING PROXY DIALER")

		// Use proxies list as failover list
		for i := 0; i < len(r.proxies); i++ {
			dialer := *r.proxies[i]

			for x, remoteAddress := range strings.Split(r.ForwardAddr, ",") {
				if r.ignoreProxySettings(x) {
					r.logger.Info.Printf("REMOTE [%d] IGNORING PROXY %d\n", x+1, i+1)

					if err := tcpDial(r, conn, remoteAddress, x, start); err != nil {
						continue
					}

					// if no error dialling then exit
					return
				}

				r.logger.Info.Printf("DIALLING FORWARD ADDRESS [%d] THROUGH PROXY %d\n", x+1, i+1)

				c, err := dialer.Dial("tcp", remoteAddress)
				if err != nil {
					r.Metrics.dial(0, 1, start)

					r.logger.Error.Printf("DIAL FORWARD ADDR: %s\n", err)
					continue
				}

				r.Metrics.dial(1, 0, start)

				r.logger.Info.Printf("CONNECTED TO %s\n", remoteAddress)
				streamConns(conn, c, r.Metrics)

				r.logger.Info.Printf("CONNECTION CLOSED %q ON %q\n", conn.RemoteAddr(), conn.LocalAddr())
				return
			}
		}

		r.logger.Error.Printf("CONNECTION CLOSED %q ON %q AFTER DIALLING WITH PROXY FAILED\n", conn.RemoteAddr(), conn.LocalAddr())
		return
	}

	// Not using proxy so dial with standard dialer
	for i, remoteAddress := range strings.Split(r.ForwardAddr, ",") {
		if err := tcpDial(r, conn, remoteAddress, i, start); err != nil {
			// dial next host
			continue
		}

		return
	}
}

func tcpDial(r *Relay, conn net.Conn, remoteAddress string, i int, start time.Time) error {
	r.logger.Info.Printf("DIALLING FORWARD ADDRESS [%d]\n", i+1)

	c, err := net.DialTimeout("tcp", remoteAddress, Timeout)
	if err != nil {
		r.Metrics.dial(0, 1, start)

		r.logger.Error.Printf("DIAL FORWARD ADDR: %s\n", err)
		return ErrFailConnect
	}

	r.Metrics.dial(1, 0, start)

	r.logger.Info.Printf("CONNECTED TO %s\n", remoteAddress)
	streamConns(conn, c, r.Metrics)

	r.logger.Info.Printf("CONNECTION CLOSED %q ON %q\n", conn.RemoteAddr(), conn.LocalAddr())
	return nil
}
