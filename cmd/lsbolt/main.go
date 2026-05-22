package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/szonov/bolt"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	if err := ls(); err != nil {
		slog.Error(err.Error())
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
		return fmt.Errorf("no receivers found")
	}

	for _, receiver := range receivers {
		if err := printReceiverInfo(receiver); err != nil {
			fmt.Printf("Receiver [%s] failed: %s\n", receiver.Name, err.Error())
		}
	}

	return nil
}

func printReceiverInfo(receiver *bolt.Receiver) error {

	if err := receiver.SetHandler(notificationHandler); err != nil {
		return fmt.Errorf("receiver.SetHandler failed: %v", err)
	}

	if err := receiver.SetSoftwareId(0x03); err != nil {
		return fmt.Errorf("receiver.SetSoftwareId failed: %v", err)
	}

	if err := receiver.Open(); err != nil {
		return fmt.Errorf("receiver.Open failed: %v", err)
	}
	defer receiver.Close()

	fmt.Printf("RECEIVER: %s (Path=%s)\n", receiver.Name, receiver.Path)

	fmt.Printf("sleep 5 seconds to see packets....")
	time.Sleep(5 * time.Second)
	return nil
}

func notificationHandler(pkt []byte) {
	fmt.Printf("notification received: %+v\n", pkt)
}
