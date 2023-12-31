package localrelay

import "net/url"

type TargetLink string

// Addr returns the address within the target link.
// Example: 127.0.0.1:443
func (t *TargetLink) Addr() string {
	u, _ := url.Parse(string(*t))
	return u.Host + ":" + u.Port()
}

// Host returns the host/ip of the target
func (t *TargetLink) Host() string {
	u, _ := url.Parse(string(*t))
	return u.Host
}

// Port returns the port number of the target
func (t *TargetLink) Port() string {
	u, _ := url.Parse(string(*t))
	return u.Port()
}

// Protocol returns the protocol of the target
func (t *TargetLink) Protocol() string {
	u, _ := url.Parse(string(*t))
	return u.Scheme
}

// TODO: TargetLink.Proxy
