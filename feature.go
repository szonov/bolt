package bolt

import "fmt"

//const (
//	FeatureDeviceName         = 0x0005
//	FeatureDeviceFriendlyName = 0x0007
//	FeatureUnifiedBattery     = 0x1004
//	FeatureChangeHost         = 0x1814
//	FeatureHostInfo           = 0x1815
//)

type Feature struct {
	Id             uint16
	Index, Version byte
	Type           TypeBitfield
}

func (f Feature) String() string {
	return fmt.Sprintf("0x%02X: 0x%04x (type: %08b, version: %d)", f.Index, f.Id, f.Type, f.Version)
}

type TypeBitfield byte
