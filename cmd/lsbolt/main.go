package main

import (
	"fmt"
	"os"

	"github.com/szonov/bolt"
)

func main() {

	if err := ls(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func ls() error {
	if err := bolt.Init(); err != nil {
		return fmt.Errorf("bolt.Init failed: %v", err)
	}
	defer bolt.Exit()

	receivers := bolt.All()
	if len(receivers) == 0 {
		fmt.Printf("no receivers found\n")
		return nil
	}

	for _, receiver := range receivers {
		if err := printReceiverInfo(receiver); err != nil {
			fmt.Printf("Receiver [%s] failed: %s\n", receiver.Name, err.Error())
		}
	}

	return nil
}

func printReceiverInfo(receiver *bolt.Receiver) error {

	fmt.Printf("\nRECEIVER: %s (Path=%s)\n", receiver.Name, receiver.Path)

	if err := receiver.Open(); err != nil {
		return fmt.Errorf("receiver.Open failed: %v", err)
	}
	defer receiver.Close()

	sn, err := receiver.GetSerialNumber()
	if err != nil {
		fmt.Printf(" -> SerialNumber: (failed to get: %v)\n", err)
	} else {
		fmt.Printf(" -> SerialNumber: '%s'\n", sn)
	}

	// bolt supports up to 6 devices
	for _, deviceIdx := range []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06} {
		device := receiver.NewDevice(deviceIdx)
		fmt.Printf(" -> Device (%d): ", deviceIdx)
		name, err := device.GetName()
		if err != nil {
			fmt.Printf("-\n")
			continue
		} else {
			fmt.Printf("%s", name)
		}
		charge, _, _, err := device.GetBatteryInfo()
		if err != nil {
			fmt.Printf(" (battery info error: %v)", err)
		} else {
			fmt.Printf(" (battery: %d%%)", charge)
		}
		fmt.Printf("\n")
	}
	return nil
}
