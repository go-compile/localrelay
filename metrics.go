package localrelay

import (
	"sync"
	"time"
)

// Metrics stores information such as bandwidth usage
// conn stats etc
type Metrics struct {
	up, down              int
	dialFail, dialSuccess uint64
	activeConns           int
	totalConns            uint64
	totalRequests         uint64

	// dialTimes holds recent durations of how long it takes a
	// relay to dial a remote
	dialTimes []int64

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

// Requests returns the amount of requests made via http
func (m *Metrics) Requests() uint64 {
	m.m.RLock()
	defer m.m.RUnlock()

	return m.totalRequests
}

// Dialer returns the successful dials and failed dials
func (m *Metrics) Dialer() (success, failed uint64) {
	m.m.RLock()
	defer m.m.RUnlock()

	return m.dialSuccess, m.dialFail
}

// DialerAvg returns the 10 point average dial time
// this average includes failed dials
func (m *Metrics) DialerAvg() (milliseconds int) {
	m.m.RLock()
	defer m.m.RUnlock()

	if len(m.dialTimes) == 0 {
		return 0
	}

	x := int64(0)
	for i := 0; i < len(m.dialTimes); i++ {
		x += m.dialTimes[i]
	}

	return int(x) / len(m.dialTimes)
}

// bandwidth will increment the bandwidth statistics
func (m *Metrics) bandwidth(up, down int) {
	m.m.Lock()
	defer m.m.Unlock()

	m.up += up
	m.down += down
}

// dial will increment the dialer success/fail statistics
func (m *Metrics) dial(success, failed uint64, t time.Time) {
	m.m.Lock()
	defer m.m.Unlock()

	m.dialSuccess += success
	m.dialFail += failed

	// 10 point moving average
	if len(m.dialTimes) >= 10 {
		m.dialTimes = append(m.dialTimes[1:], time.Since(t).Milliseconds())
	} else {
		m.dialTimes = append(m.dialTimes, time.Since(t).Milliseconds())
	}
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

// requests will update the requests metric
func (m *Metrics) requests(delta int) {
	m.m.Lock()
	defer m.m.Unlock()

	m.totalRequests += uint64(delta)
}
