package webserver

import (
	"net"
	"os"
	"syscall"
)

func FileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	return err == nil && !info.IsDir()
}

func checkError(err error) {
	if err == nil {
		return
	}

	// Check for a broken connection, as it is not really a
	// condition that warrants a panic stack trace.
	if ne, ok := err.(*net.OpError); ok {
		if se, ok := ne.Err.(*os.SyscallError); ok {
			if se.Err == syscall.EPIPE || se.Err == syscall.ECONNRESET {
				return
			}
		}
	}

	panic(err)
}
