// +build !windows

package main

import (
	"golang.org/x/sys/unix"
	"os"
)

func KillProcess(id int) error {
	p, err := os.FindProcess(id)
	if err != nil {
		return err
	}
	if err == nil && p != nil {
		return p.Signal(unix.SIGINT)
	}
	return nil
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
