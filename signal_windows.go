package main

import (
	"golang.org/x/sys/windows"
	"os"
)

func KillProcess(id int) error {
	p, err := os.FindProcess(id)
	if err != nil {
		return err
	}
	if err == nil && p != nil {
		return p.Signal(windows.SIGINT)
	}
	return nil
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
