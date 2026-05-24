package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/szonov/bolt"
)

func cmdMonitor() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	if err := cmdMonitorHandler(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func cmdMonitorHandler() error {
	if err := bolt.Init(); err != nil {
		return fmt.Errorf("bolt.Init failed: %v", err)
	}
	defer bolt.Exit()

	receiver := bolt.First()

	if receiver == nil {
		fmt.Printf("no receivers found\n")
		return nil
	}
	fmt.Printf("\nRECEIVER: %s (Path=%s)\n", receiver.Name, receiver.Path)

	if err := receiver.Open(); err != nil {
		return fmt.Errorf("receiver.Open failed: %v", err)
	}
	defer receiver.Close()

	fmt.Printf("CTRL-C to quit\n")

	waitForSignal()

	return nil
}
