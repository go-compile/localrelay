package ipc

import (
	"net"

	"gopkg.in/natefinch/npipe.v2"
)

// NewListener for windows uses name pipes to communicate
func NewListener() (net.Listener, error) {
	return npipe.Listen(`\\.\pipe\` + serviceName)
}
