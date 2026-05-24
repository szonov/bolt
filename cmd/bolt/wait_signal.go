package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

// waitForSignal wait for os.Signal and run callbacks:
//   - first - callback for exit signal
//   - second - callback for SIGUSR1
//   - third - callback for SIGUSR2
func waitForSignal(cb ...func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c,
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGABRT,
		syscall.SIGTERM,
		syscall.SIGUSR1,
		syscall.SIGUSR2,
	)

	count := len(cb)

	for {
		signalType := <-c
		slog.Info("[os.Signal]", "signal", signalType)
		switch signalType {
		case syscall.SIGUSR1:
			// Handle SIGUSR1 signal
			// e.g., reload configuration, reset state, etc.
			// example: kill -USR1 <PID>
			if count > 1 {
				cb[1]()
			}
		case syscall.SIGUSR2:
			// Handle SIGUSR2 signal
			// e.g., reload configuration, reset state, etc.
			// example: kill -USR2 <PID>
			if count > 2 {
				cb[2]()
			}
		default:
			if count > 0 {
				cb[0]()
			}
			return
		}
	}
}
