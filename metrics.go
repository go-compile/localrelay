package localrelay

import "sync"

// Metrics stores information such as bandwidth usage
// conn stats etc
type Metrics struct {
	up, down              int
	dialFail, dialSuccess uint64
	activeConns           uint16

	m sync.RWMutex
}

// Upload returns the amount of bytes uploaded through the relay
func (m *Metrics) Upload() int {
	m.m.RLock()
	defer m.m.RUnlock()

	return m.up
}

// Download returns the amount of bytes downloaded through the relay
func (m *Metrics) Download() int {
	m.m.RLock()
	defer m.m.RUnlock()

	return m.down
}

// bandwidth will increment the bandwidth statistics
func (m *Metrics) bandwidth(up, down int) {
	m.m.Lock()
	defer m.m.Unlock()

	m.up += up
	m.down += down
}
