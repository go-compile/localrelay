package localrelay

import (
	"errors"
	"io"
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

		streamConns(conn, c)

		return
	}

	// Not using proxy so dial with standard dialer

	r.logger.Info.Println("DIALLING FORWARD ADDRESS")
	c, err := net.Dial("tcp", r.ForwardAddr)
	if err != nil {
		r.logger.Error.Printf("DIAL FORWARD ADDR: %s\n", err)
		return
	}

	streamConns(conn, c)
}

func streamConns(client net.Conn, remote net.Conn) {
	go io.Copy(client, remote)
	io.Copy(remote, client)

	remote.Close()
	client.Close()
}
