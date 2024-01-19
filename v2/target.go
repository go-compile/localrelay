package localrelay

import (
	"errors"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/mroth/weightedrand"
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
func (t *TargetLink) Proxy(r *Relay) ([]ProxyURL, []string, error) {
	u, _ := url.Parse(string(*t))

	// get ?proxy=<value> from TargetLink and split into comma seperated array
	proxieNames := strings.Split(u.Query().Get("proxy"), ",")
	if len(proxieNames) == 0 || len(proxieNames[0]) == 0 {
		return nil, nil, nil
	}

	proxies := make([]ProxyURL, len(proxieNames))
	for i := 0; i < len(proxies); i++ {
		proxy, found := r.proxies[proxieNames[i]]
		if !found {
			return proxies, proxieNames, ErrProxyDefine
		}

		proxies[i] = proxy
	}

	return proxies, proxieNames, nil
}

// Print returns the targetlink string
func (t *TargetLink) Print() string {
	return string(*t)
}

// LbWeight returns the weight provided or the default value of 100
func (t *TargetLink) LbWeight() uint {
	u, _ := url.Parse(string(*t))
	weight := u.Query().Get("lb_weight")

	if weight == "" {
		return 100
	}

	n, err := strconv.Atoi(weight)
	if err != nil {
		log.Fatalf("Weight could not be parsed for %s\n", t)
	}

	return uint(n)
}

// Lb returns true if loadbalancing is enabled
func (t *TargetLink) Lb() bool {
	u, _ := url.Parse(string(*t))
	switch strings.ToLower(u.Query().Get("lb")) {
	case "false", "off", "disabled", "inactive", "0", "no":
		return false
	default:
		return true
	}
}

// nextDestination using the provided list of potential destinations, find the appripriate
// next one to try based on the relay config, e.g. loadbalance and failovers.
func nextDestination(r *Relay, dsts []TargetLink) (int, TargetLink, error) {
	candiates := []TargetLink{}

	if r.loadbalance.Enabled {
		choices := []weightedrand.Choice{}
		// Remove all non loadbalanced dsts
		for i := 0; i < len(dsts); i++ {
			if dsts[i].Lb() {
				candiates = append(candiates, dsts[i])
				choices = append(choices, weightedrand.NewChoice(i, dsts[i].LbWeight()))
			}
		}

		// there are no load balanced relays left, use the failovers
		if len(candiates) == 0 {
			return 0, dsts[0], nil
		}

		chooser, err := weightedrand.NewChooser(choices...)
		if err != nil {
			return 0, "", err
		}

		dstI := chooser.Pick().(int)

		return dstI, dsts[dstI], nil
	}

	// return the first destination
	return 0, dsts[0], nil
}
