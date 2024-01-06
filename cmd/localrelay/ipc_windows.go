package main

import (
	"os"

	"golang.org/x/sys/windows"
)

func fileOwnership(stat os.FileInfo) (string, error) {
	// TODO: get owner of file on windows
	return "", nil
}

func runningAsRoot() bool {
	token := windows.GetCurrentProcessToken()
	defer token.Close()

	return token.IsElevated()
}
