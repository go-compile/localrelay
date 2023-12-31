package localrelay

import (
	"net/url"
	"strings"
)

type TargetLink string

func (t *TargetLink) String() string {
	u, _ := url.Parse(string(*t))
	return u.Host
}

// ProxyType returns the protocol as a ProxyType
func (t *TargetLink) ProxyType() ProxyType {
	return ProxyType(t.Protocol())
}

// Addr returns the address within the target link.
// Example: 127.0.0.1:443
func (t *TargetLink) Addr() string {
	u, _ := url.Parse(string(*t))
	return u.Host
}

// Host returns the host/ip of the target
func (t *TargetLink) Host() string {
	u, _ := url.Parse(string(*t))
	return u.Hostname()
}

// Port returns the port number of the target
func (t *TargetLink) Port() string {
	u, _ := url.Parse(string(*t))

	if len(u.Port()) > 0 {
		return u.Port()
	}
	switch t.Protocol() {
	case "https":
		return "443"
	case "http":
		return "80"
	default:
		return ""
	}
}

// Protocol returns the protocol of the target
func (t *TargetLink) Protocol() string {
	u, _ := url.Parse(string(*t))
	return strings.ToLower(u.Scheme)
}

// TODO: TargetLink.Proxy
