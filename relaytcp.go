package localrelay

import (
	"io"
	"net"
	"sync"
	"time"

	"github.com/pkg/errors"
)

func listener(r *Relay) (net.Listener, error) {

	network := "tcp"
	if r.ProxyType == ProxyUDP {
		network = "udp"
	}

	l, err := net.Listen(network, r.Host)
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

		go handleConn(r, conn, "tcp")
	}
}

func handleConn(r *Relay, conn net.Conn, network string) {
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

			err = streamConns(conn, c, r.Metrics)
			if err != nil {
				r.logger.Info.Printf("ERROR FROM %q ON %q: ERR=%s\n", conn.RemoteAddr(), conn.LocalAddr(), err)
			}

			r.logger.Info.Printf("CONNECTION CLOSED %q ON %q\n", conn.RemoteAddr(), conn.LocalAddr())
			return
		}

		r.logger.Error.Printf("CONNECTION CLOSED %q ON %q AFTER DIALLING WITH PROXY FAILED\n", conn.RemoteAddr(), conn.LocalAddr())
		return
	}

	// Not using proxy so dial with standard dialer

	r.logger.Info.Println("DIALLING FORWARD ADDRESS")

	proto := network

	// if protocol switching enabled set the new network
	if protocol, ok := r.protocolSwitching[0]; ok {
		r.logger.Info.Printf("SWITCHING PROTOCOL FROM %q TO %q\n", network, protocol)
		proto = protocol
	}

	c, err := net.DialTimeout(proto, r.ForwardAddr, Timeout)
	if err != nil {
		r.Metrics.dial(0, 1, start)

		r.logger.Error.Printf("DIAL FORWARD ADDR: %s\n", err)
		return
	}

	r.Metrics.dial(1, 0, start)

	r.logger.Info.Printf("CONNECTED TO %s\n", r.ForwardAddr)
	err = streamConns(conn, c, r.Metrics)
	if err != nil {
		r.logger.Info.Printf("ERROR FROM %q ON %q: ERR=%s\n", conn.RemoteAddr(), conn.LocalAddr(), err)
	}

	r.logger.Info.Printf("CONNECTION CLOSED %q ON %q\n", conn.RemoteAddr(), conn.LocalAddr())
}

func streamConns(client net.Conn, remote net.Conn, m *Metrics) error {
	wg := sync.WaitGroup{}

	var copyInErr error

	wg.Add(1)
	go func() {
		copyInErr = copierIn(client, remote, 128, m)
		wg.Done()
	}()

	wg.Add(1)
	err := copierOut(client, remote, 128, m)
	wg.Done()

	// if error is reporting that the conn is closed ignore both
	if errors.Is(copyInErr, io.EOF) || errors.Is(err, io.EOF) || errors.Is(copyInErr, net.ErrClosed) {
		// if any of the errors are EOFs or ErrClosed we are not
		//  bothered with any additional errors.
		return nil
	}

	// else propagate error
	return err
}

// NOTE: static function for maximum performance
func copierIn(client net.Conn, dst net.Conn, buffer int, m *Metrics) error {

	buf := make([]byte, buffer)
	for {
		n, err := dst.Read(buf)
		m.bandwidth(0, n)
		if err != nil {

			var err1 error
			// if we read some data, flush it then return a error
			if n > 0 {
				_, err1 = dst.Write(buf[:n])
			}

			err2 := client.Close()
			err3 := dst.Close()

			// wrap all errors into one
			if err1 != nil {
				err = errors.Wrap(err, err1.Error())
			}

			// wrap all errors into one
			if err2 != nil {
				err = errors.Wrap(err, err2.Error())
			}

			// wrap all errors into one
			if err3 != nil {
				err = errors.Wrap(err, err3.Error())
			}

			return errors.WithStack(err)
		}

		if n2, err := client.Write(buf[:n]); err != nil || n2 != n {
			err1 := client.Close()
			err2 := dst.Close()

			// wrap all errors into one
			if err1 != nil {
				err = errors.Wrap(err, err1.Error())
			}

			// wrap all errors into one
			if err2 != nil {
				err = errors.Wrap(err, err2.Error())
			}

			return errors.WithStack(err)
		}
	}
}

// NOTE: static function for maximum performance
func copierOut(client net.Conn, dst net.Conn, buffer int, m *Metrics) error {

	buf := make([]byte, buffer)
	for {

		n, err := client.Read(buf)
		m.bandwidth(n, 0)
		if err != nil {

			var err1 error
			// if we read some data, flush it then return a error
			if n > 0 {
				_, err1 = dst.Write(buf[:n])
			}

			err2 := client.Close()
			err3 := dst.Close()

			// wrap all errors into one
			if err1 != nil {
				err = errors.Wrap(err, err1.Error())
			}

			// wrap all errors into one
			if err2 != nil {
				err = errors.Wrap(err, err2.Error())
			}

			// wrap all errors into one
			if err3 != nil {
				err = errors.Wrap(err, err3.Error())
			}

			return errors.WithStack(err)
		}

		if n2, err := dst.Write(buf[:n]); err != nil || n2 != n {
			err1 := client.Close()
			err2 := dst.Close()

			// wrap all errors into one
			if err1 != nil {
				err = errors.Wrap(err, err1.Error())
			}

			// wrap all errors into one
			if err2 != nil {
				err = errors.Wrap(err, err2.Error())
			}

			return errors.WithStack(err)
		}
	}
}
