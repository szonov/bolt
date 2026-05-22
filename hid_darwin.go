//go:build darwin

package bolt

import "github.com/sstallion/go-hid"

func disableHidExclusive() {
	hid.SetOpenExclusive(false)
}
