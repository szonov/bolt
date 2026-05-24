package main

import (
	"os"
	"path/filepath"
)

type Command struct {
	name  string
	usage string
}

var programName string
var version = "0.0.1"
var commands = []Command{
	{"list", "Display bolt receivers with attached devices"},
	{"monitor", "Monitor events from bolt receiver"},
	{"help", "Display this help"},
	{"version", "Display program version"},
}

func main() {
	programName = filepath.Base(os.Args[0])
	var cmd string

	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	switch cmd {
	case "-v", "--version", "version":
		cmdVersion()
	case "list":
		cmdList()
	case "monitor":
		cmdMonitor()
	default:
		cmdHelp()
	}
}
