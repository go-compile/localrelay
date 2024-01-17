//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || !windows
// +build darwin dragonfly freebsd linux netbsd openbsd solaris !windows

package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
)

var (
	// ipcPathPrefix is the dir which comes before the unix socket
	ipcPathPrefix = "/var/run/"
)

type logger struct {
	w         io.WriteCloser
	relayName string
}

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
	return os.Geteuid() == 0
}

func (l *logger) Write(b []byte) (int, error) {
	return l.w.Write(b)
}

func (l *logger) Close() error {
	return l.w.Close()
}

func newLogger(relayName string) *logger {
	f, err := os.OpenFile(filepath.Join("/var/log/localrelay/", relayName+".log"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}

	return &logger{
		w:         f,
		relayName: relayName,
	}
}
