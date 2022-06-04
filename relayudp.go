package localrelay

import (
	"errors"
	"net"
)

func relayUDP(r *Relay, l net.Listener) error {

	r.logger.Info.Println("STARTING UDP RELAY")

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

		go handleConn(r, conn, "udp")
	}
}
