package bolt

import (
	"fmt"
)

const (
	FeatureDeviceName         = 0x0005
	FeatureDeviceFriendlyName = 0x0007
	FeatureUnifiedBattery     = 0x1004
	FeatureChangeHost         = 0x1814
	FeatureHostInfo           = 0x1815
)

func (d *Device) GetName() (string, error) {
	return d.getText(FeatureDeviceName, 0)
}

func (d *Device) GetFriendlyName() (string, error) {
	return d.getText(FeatureDeviceFriendlyName, 1)
}

func (d *Device) GetBatteryInfo() (charge int, level, status byte, err error) {
	var b []byte
	if b, _, err = d.FeatureRequest(FeatureUnifiedBattery, 1); err != nil {
		err = fmt.Errorf("GetBatteryInfo(0x%04X) failed to get info: %w", FeatureUnifiedBattery, err)
		return
	}

	charge = int(b[4])
	level = b[5]
	status = b[6]

	return
}

func (d *Device) getText(featureId uint16, textOffset int) (string, error) {
	// function 0 = GetCount()
	b, featureIndex, err := d.FeatureRequest(featureId, 0)
	if err != nil {
		return "", fmt.Errorf("getText(0x%04X) failed to get length: %w", featureId, err)
	}

	length := int(b[4])

	name := make([]byte, 0, length)

	for offset := 0; offset < length; {
		// function 1 = GetName(offset)
		if b, err = d.Request(featureIndex, 1, byte(offset)); err != nil {
			return "", fmt.Errorf(
				"getText(%02X) offset %d failed: %w",
				featureIndex,
				offset,
				err,
			)
		}

		chunk := b[4+textOffset:]

		remaining := length - offset

		if len(chunk) > remaining {
			chunk = chunk[:remaining]
		}

		name = append(name, chunk...)

		offset += len(chunk)
	}

	return string(name), nil
}
