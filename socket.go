package main

import (
	"net"
	"syscall"
)

func getPidOfConn(conn *net.UnixConn) (int32, error) {
	f, err := conn.File()
	if err != nil {
		return 0, err
	}
	defer f.Close()

	cred, err := syscall.GetsockoptUcred(int(f.Fd()), syscall.SOL_SOCKET, syscall.SO_PEERCRED)
	if err != nil {
		return 0, err
	}

	return cred.Pid, nil
}
