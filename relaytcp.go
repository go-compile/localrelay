package localrelay

import (
	"errors"
	"net"
	"time"
)

func listener(r *Relay) (net.Listener, error) {
	l, err := net.Listen("tcp", r.Host)
	if err != nil {
		return nil, err
	}

	return l, nil
}

func relayTCP(r *Relay, l net.Listener) error {

	r.logger.Info.Println("STARTING TCP RELAY")

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

		go handleTCP(r, conn)
	}
}

func handleTCP(r *Relay, conn net.Conn) {
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

			r.logger.Info.Printf("DIALLING FORWARD ADDRESS THROUGH PROXY %d\n", i+1)

			c, err := dialer.Dial("tcp", r.ForwardAddr)
			if err != nil {
				r.Metrics.dial(0, 1, start)

				r.logger.Error.Printf("DIAL FORWARD ADDR: %s\n", err)
				continue
			}

			r.Metrics.dial(1, 0, start)

			r.logger.Info.Printf("CONNECTED TO %s\n", r.ForwardAddr)
			streamConns(conn, c, r.Metrics)

			r.logger.Info.Printf("CONNECTION CLOSED %q ON %q\n", conn.RemoteAddr(), conn.LocalAddr())
			return
		}

		r.logger.Error.Printf("CONNECTION CLOSED %q ON %q AFTER DIALLING WITH PROXY FAILED\n", conn.RemoteAddr(), conn.LocalAddr())
		return
	}

	// Not using proxy so dial with standard dialer

	r.logger.Info.Println("DIALLING FORWARD ADDRESS")
	c, err := net.DialTimeout("tcp", r.ForwardAddr, Timeout)
	if err != nil {
		r.Metrics.dial(0, 1, start)

		r.logger.Error.Printf("DIAL FORWARD ADDR: %s\n", err)
		return
	}

	r.Metrics.dial(1, 0, start)

	r.logger.Info.Printf("CONNECTED TO %s\n", r.ForwardAddr)
	streamConns(conn, c, r.Metrics)

	r.logger.Info.Printf("CONNECTION CLOSED %q ON %q\n", conn.RemoteAddr(), conn.LocalAddr())
}

func streamConns(client net.Conn, remote net.Conn, m *Metrics) {
	go copierIn(client, remote, 128, m)
	copierOut(remote, client, 128, m)
}

// NOTE: statics function for maximum performance
func copierIn(src net.Conn, dst net.Conn, buffer int, m *Metrics) error {

	buf := make([]byte, buffer)
	for {

		n, err := src.Read(buf)
		m.bandwidth(n, 0)
		if err != nil {

			// if we read some data, flush it then return a error
			if n > 0 {
				dst.Write(buf[:n])
			}

			src.Close()
			dst.Close()

			return err
		}

		if n2, err := dst.Write(buf[:n]); err != nil || n2 != n {
			src.Close()
			dst.Close()

			return err
		}
	}
}

// NOTE: statics function for maximum performance
func copierOut(src net.Conn, dst net.Conn, buffer int, m *Metrics) error {

	buf := make([]byte, buffer)
	for {

		n, err := src.Read(buf)
		m.bandwidth(0, n)
		if err != nil {

			// if we read some data, flush it then return a error
			if n > 0 {
				dst.Write(buf[:n])
			}

			src.Close()
			dst.Close()

			return err
		}

		if n2, err := dst.Write(buf[:n]); err != nil || n2 != n {
			src.Close()
			dst.Close()

			return err
		}
	}
}
