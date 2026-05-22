package bolt

import (
	"github.com/sstallion/go-hid"
)

func Init() error {
	if err := hid.Init(); err != nil {
		return err
	}
	disableHidExclusive()
	return nil
}

func Exit() error {
	return hid.Exit()
}
