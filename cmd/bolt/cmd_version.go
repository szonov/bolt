package main

import "fmt"

func cmdVersion() {
	fmt.Printf("%s version: %s\n", programName, version)
}
