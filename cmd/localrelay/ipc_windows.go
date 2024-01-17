package main

import (
	"log"
	"os"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/eventlog"
)

type logger struct {
	w         *eventlog.Log
	relayName string
}

func fileOwnership(stat os.FileInfo) (string, error) {
	// TODO: get owner of file on windows
	return "", nil
}

func runningAsRoot() bool {
	token := windows.GetCurrentProcessToken()
	defer token.Close()

	return token.IsElevated()
}

func (l *logger) Write(b []byte) (int, error) {
	return len(b), l.w.Info(1, string(b))
}

func (l *logger) Close() error {
	return l.w.Close()
}

func newLogger(relayName string) *logger {
	w, err := eventlog.Open("localrelayd")
	if err != nil {
		log.Fatal(err)
	}

	return &logger{
		w:         w,
		relayName: relayName,
	}
}
