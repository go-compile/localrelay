package localrelay

import (
	"net"
)

func defaultDialer() net.Dialer {
	return net.Dialer{
		DualStack: false,
		KeepAlive: 15,
	}
}
