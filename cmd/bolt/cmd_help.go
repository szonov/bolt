package main

import "fmt"

func cmdHelp() {
	fmt.Printf("%s - program for interaction with Logitech Bolt Receiver\n\n", programName)
	fmt.Printf("Usage: %s COMMAND [OPTIONS]\n\n", programName)
	fmt.Printf("Available commands:\n")

	for _, cmd := range commands {
		fmt.Printf("  %-20s %s\n", cmd.name, cmd.usage)
	}
}
