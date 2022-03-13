package localrelay

import (
	"errors"
	"net"
)

func listenerTCP(r *Relay) (net.Listener, error) {
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
	defer conn.Close()

	r.logger.Info.Printf("NEW CONNECTION %q ON %q\n", conn.RemoteAddr(), conn.LocalAddr())

	// If using a proxy dial with proxy
	if r.proxy != nil {
		r.logger.Info.Println("CREATING PROXY DIALER")
		dialer := *r.proxy

		r.logger.Info.Println("DIALLING FORWARD ADDRESS THROUGH PROXY")
		c, err := dialer.Dial("tcp", r.ForwardAddr)
		if err != nil {
			r.logger.Error.Printf("DIAL FORWARD ADDR: %s\n", err)
			return
		}

		r.logger.Info.Printf("CONNECTED TO %s\n", r.ForwardAddr)
		streamConns(conn, c)

		r.logger.Info.Printf("CONNECTION CLOSED %q ON %q\n", conn.RemoteAddr(), conn.LocalAddr())
		return
	}

	// Not using proxy so dial with standard dialer

	r.logger.Info.Println("DIALLING FORWARD ADDRESS")
	c, err := net.Dial("tcp", r.ForwardAddr)
	if err != nil {
		r.logger.Error.Printf("DIAL FORWARD ADDR: %s\n", err)
		return
	}

	r.logger.Info.Printf("CONNECTED TO %s\n", r.ForwardAddr)
	streamConns(conn, c)

	r.logger.Info.Printf("CONNECTION CLOSED %q ON %q\n", conn.RemoteAddr(), conn.LocalAddr())
}

func streamConns(client net.Conn, remote net.Conn) {
	go copier(client, remote, 128)
	copier(remote, client, 128)
}

func copier(src net.Conn, dst net.Conn, buffer int) error {

	buf := make([]byte, buffer)
	for {

		n, err := src.Read(buf)
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
