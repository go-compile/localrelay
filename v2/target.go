package localrelay

import (
	"errors"
	"net/url"
	"strings"

	"golang.org/x/net/proxy"
)

var (
	ErrProxyDefine = errors.New("proxy is not defined")
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
	return u.Hostname() + ":" + t.Port()
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

// Proxy parses the TargetLink and uses the relay to lookup proxy dialers.
// The returned array is in the same order as written.
func (t *TargetLink) Proxy(r *Relay) ([]proxy.Dialer, []string, error) {
	u, _ := url.Parse(string(*t))

	// get ?proxy=<value> from TargetLink and split into comma seperated array
	proxieNames := strings.Split(u.Query().Get("proxy"), ",")
	if len(proxieNames) == 0 {
		return nil, proxieNames, nil
	}

	proxies := make([]proxy.Dialer, len(proxieNames))
	for i := 0; i < len(proxies); i++ {
		proxy, found := r.proxies[proxieNames[i]]
		if !found {
			return proxies, proxieNames, ErrProxyDefine
		}

		proxies[i] = proxy
	}

	return proxies, proxieNames, nil
}
