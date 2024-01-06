//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || !windows
// +build darwin dragonfly freebsd linux netbsd openbsd solaris !windows

package main

import (
	"os"
	"strconv"
	"syscall"
)

var (
	// ipcPathPrefix is the dir which comes before the unix socket
	ipcPathPrefix = "/var/run/"
)

func fileOwnership(stat os.FileInfo) (string, error) {
	s := stat.Sys().(*syscall.Stat_t)
	uid := s.Uid
	gid := s.Gid

	if uid != gid {
		return strconv.Itoa(int(uid)) + "," + strconv.Itoa(int(gid)), nil
	}

	return strconv.Itoa(int(uid)), nil
}

func runningAsRoot() bool {
	return os.Geteuid() == 1
}
