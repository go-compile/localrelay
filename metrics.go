package localrelay

import "sync"

// Metrics stores information such as bandwidth usage
// conn stats etc
type Metrics struct {
	up, down              int
	dialFail, dialSuccess uint64
	activeConns           int
	totalConns            uint64

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

// Connections returns the amount of active and total connections
func (m *Metrics) Connections() (active int, total uint64) {
	m.m.RLock()
	defer m.m.RUnlock()

	return m.activeConns, m.totalConns
}

// Dialer returns the successful dials and failed dials
func (m *Metrics) Dialer() (success, failed uint64) {
	m.m.RLock()
	defer m.m.RUnlock()

	return m.dialSuccess, m.dialFail
}

// bandwidth will increment the bandwidth statistics
func (m *Metrics) bandwidth(up, down int) {
	m.m.Lock()
	defer m.m.Unlock()

	m.up += up
	m.down += down
}

// dial will increment the dialer success/fail statistics
func (m *Metrics) dial(success, failed uint64) {
	m.m.Lock()
	defer m.m.Unlock()

	m.dialSuccess += success
	m.dialFail += failed
}

// connections will update the active connections metric
func (m *Metrics) connections(delta int) {
	m.m.Lock()
	defer m.m.Unlock()

	// Calculate total connections
	if delta > 0 {
		m.totalConns += uint64(delta)
	}

	m.activeConns += delta
}
