// +build !windows

package main

import (
	"golang.org/x/sys/unix"
	"os"
)

func KillProcess(id int) {
	p, err := os.FindProcess(id)
	if err == nil && p != nil {
		_ = p.Signal(unix.SIGINT)
	}
}

func GetSingal() []os.Signal {
	return []os.Signal{
		os.Interrupt,
		os.Kill,
		unix.SIGHUP,
		//unix.SIGCHLD,
		unix.SIGKILL,
		unix.SIGINT,
		unix.SIGTERM,
		unix.SIGQUIT,
	}
}
