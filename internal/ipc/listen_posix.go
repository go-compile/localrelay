//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || !windows
// +build darwin dragonfly freebsd linux netbsd openbsd solaris !windows

package ipc

import "net"

// NewListener for windows uses name pipes to communicate
func NewListener() (net.Listener, error) {
	return net.Listen("unix", ipcPathPrefix+ipcSocket)
}
