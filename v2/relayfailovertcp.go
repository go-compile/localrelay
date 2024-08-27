package localrelay

import (
	"errors"
	"net"
	"time"
)

var (
	// ErrFailConnect will be returned if the remote failed to dial
	ErrFailConnect = errors.New("failed to dial remote")
	// Timeout is only used when dialling without a proxy
	Timeout = time.Second * 5
)

func dial(r *Relay, conn net.Conn, remoteAddress string, i int, network string, start time.Time) error {
	r.logger.Info.Printf("DIALLING FORWARD ADDRESS [%d]\n", i+1)

	c, err := net.DialTimeout(network, remoteAddress, Timeout)
	if err != nil {
		r.Metrics.dial(0, 1, start)

		r.logger.Error.Printf("DIAL FORWARD ADDR: %s\n", err)
		return ErrFailConnect
	}

	r.setConnRemote(conn, c.RemoteAddr())

	r.Metrics.dial(1, 0, start)

	r.logger.Info.Printf("CONNECTED TO %s\n", remoteAddress)
	streamConns(conn, c, r.Metrics)

	r.logger.Info.Printf("CONNECTION CLOSED %q ON %q\n", conn.RemoteAddr(), conn.LocalAddr())
	return nil
}
