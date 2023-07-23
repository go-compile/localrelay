package main

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/windows"
	"gopkg.in/natefinch/npipe.v2"
)

// fork requests elevated privileges via UAC then forks the process and provides an
// IPC pipe to communicate back to the master process.
func fork() error {
	// create a IPC listener from the unprivileged process
	connCh := make(chan net.Conn)
	pipe, l, err := createTmpIPC(connCh)
	if err != nil {
		return err
	}

	// close IPC listener
	defer l.Close()

	// request UAC to create new process and provide IPC pipe
	if err := elevatePrivileges(append(os.Args[1:], []string{"-ipc-stream-io-pipe", pipe}...)); err != nil {
		return err
	}

	// create new IPC timeout
	timeout := time.NewTicker(time.Second * 30)

	select {
	case <-timeout.C:
		timeout.Stop()
		return ErrIPCTimeout
	case conn := <-connCh:
		timeout.Stop()

		defer conn.Close()

		// stream elevated's stdout to us and our stdin to elevated process
		go io.Copy(conn, os.Stdin)
		io.Copy(os.Stdout, conn)
	}

	return nil
}

// elevatePrivileges
func elevatePrivileges(args []string) error {
	verb := "runas"
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	cmdArgs := strings.Join(args, " ")

	verbPtr, _ := syscall.UTF16PtrFromString(verb)
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)
	argPtr, _ := syscall.UTF16PtrFromString(cmdArgs)

	showCmd := int32(0)
	return windows.ShellExecute(0, verbPtr, exePtr, argPtr, cwdPtr, showCmd)
}

// createTmpIPC is used to forward io from a newly connected process to an existing one
func createTmpIPC(connCh chan net.Conn) (string, io.Closer, error) {
	randBuf := make([]byte, 16)
	_, err := rand.Read(randBuf)
	if err != nil {
		return "", nil, err
	}

	pipe := `\\.\pipe\` + "localrelay-stream." + hex.EncodeToString(randBuf)

	// create a new name pipe with a unique name
	l, err := npipe.Listen(pipe)
	if err != nil {
		return pipe, l, err
	}

	// asynchronously wait for a connection then push to channel
	go func(l *npipe.PipeListener, connCh chan net.Conn) {
		conn, _ := l.Accept()

		connCh <- conn
	}(l, connCh)

	return pipe, l, nil
}

func forwardIO(opt *options) (net.Conn, error) {
	conn, err := npipe.DialTimeout(opt.ipcPipe, time.Second*5)
	if err != nil {
		return nil, err
	}

	stdout = conn
	log.SetOutput(conn)

	return conn, nil
}
