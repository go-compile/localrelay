package localrelay

import (
	"net"
)

func relayHTTPS(r *Relay, l net.Listener) error {
	r.logger.Info.Println("STARTING HTTPS RELAY")

	return r.server.ServeTLS(l, r.certificateFile, r.keyFile)
}
