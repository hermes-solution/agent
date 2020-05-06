package main

import (
	"golang.org/x/sys/windows"
	"os"
)

func KillProcess(id int) {
	p, err := os.FindProcess(id)
	if err == nil && p != nil {
		_ = p.Signal(windows.SIGINT)
	}
}

func GetSingal() []os.Signal {
	return []os.Signal{
		os.Interrupt,
		os.Kill,
		windows.SIGHUP,
		windows.SIGKILL,
		windows.SIGINT,
		windows.SIGTERM,
		windows.SIGQUIT,
	}
}
