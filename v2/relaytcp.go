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
	if r.Listener.ProxyType() == ProxyUDP {
		network = "udp"
	}

	l, err := net.Listen(network, r.Listener.Addr())
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
	r.storeConn(conn)

	defer func() {
		conn.Close()

		// remove conn from connPool
		r.popConn(conn)

		r.Metrics.connections(-1)
	}()

	r.Metrics.connections(1)

	r.logger.Info.Printf("NEW CONNECTION %q ON %q\n", conn.RemoteAddr(), conn.LocalAddr())

	start := time.Now()

	for i, destination := range r.Destination {
		proxies, proxyNames, err := destination.Proxy(r)
		if err != nil {
			r.logger.Error.Printf("A PROXY FOR DESTINATION %q WAS REFERENCED BUT NOT DEFINED\n", destination)
			return
		}

		// if no proxy is set
		if proxies == nil {
			r.logger.Info.Printf("DAILING REMOTE [%s]\n", destination)

			if err := dial(r, conn, destination.Addr(), i+1, destination.Protocol(), start); err != nil {
				r.logger.Info.Printf("FAILED DAILING REMOTE [%s]\n", destination)
				// errored dialing, continue to try next destination
				continue
			}

			r.logger.Info.Printf("CONNECTION CLOSED %q ON %q\n", conn.RemoteAddr(), conn.LocalAddr())
			// close connection
			return
		}

		// proxies are set for this destination
		for pi, proxy := range proxies {
			r.logger.Info.Printf("DIALLING DESTINATION [%d] ADDRESS [%s] THROUGH PROXY %q\n", i+1, destination, proxyNames[pi])

			// Dial destination through proxy
			c, err := proxy.Dial(destination.Protocol(), destination.Addr())
			if err != nil {
				r.Metrics.dial(0, 1, start)

				r.logger.Error.Printf("FAILED TO DIAL DESTINATION ADDR: %s\n", err)
				// try next proxy
				continue
			}

			r.Metrics.dial(1, 0, start)

			r.logger.Info.Printf("CONNECTED TO %s\n", destination)
			streamConns(conn, c, r.Metrics)

			r.logger.Info.Printf("CONNECTION CLOSED %q ON %q\n", conn.RemoteAddr(), conn.LocalAddr())
			// close connection
			return
		}
	}

	r.logger.Info.Printf("UNABLE TO MAKE A CONNECTION FROM %q TO %q\n", conn.RemoteAddr(), conn.LocalAddr())
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
