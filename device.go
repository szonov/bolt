package bolt

import (
	"fmt"
	"sync"
)

type Device struct {
	receiver *Receiver
	index    byte
	features map[uint16]byte
	mu       sync.RWMutex
}

func NewDevice(receiver *Receiver, index byte) *Device {
	return &Device{
		receiver: receiver,
		index:    index,
		features: make(map[uint16]byte),
	}
}

func (d *Device) Index() byte {
	return d.index
}

func (d *Device) SetFeatureIndex(featureId uint16, featureIndex byte) {
	d.mu.Lock()
	d.features[featureId] = featureIndex
	d.mu.Unlock()
}

func (d *Device) GetFeatureIndex(featureId uint16) (featureIndex byte, found bool) {
	d.mu.RLock()
	featureIndex, found = d.features[featureId]
	d.mu.RUnlock()

	if !found {
		// если нет в нашей мапе, то обращаемся к устройству - пусть оно скажет
		// по какому адресу у него эта фича расположена
		featureIndex, found = d.checkFeature(featureId)
	}

	if featureIndex == 0x00 {
		found = false
	}

	return
}

func (d *Device) checkFeature(featureId uint16) (featureIndex byte, found bool) {
	b, err := d.Request(0, 0, byte(featureId>>8), byte(featureId))

	if err == nil && len(b) > 4 && b[4] > 0 {
		featureIndex = b[4]
		found = true
	} else {
		featureIndex = 0x00
	}

	d.SetFeatureIndex(featureId, featureIndex)

	return
}

func (d *Device) CheckFeature(featureId uint16) (featureIndex byte, found bool) {
	b, err := d.Request(0, 0, byte(featureId>>8), byte(featureId))

	if err == nil && b[4] > 0 {
		featureIndex = b[4]
		found = true
		d.SetFeatureIndex(featureId, featureIndex)
	}

	return
}

func (d *Device) Request(featureIndex, functionIndex byte, data ...byte) ([]byte, error) {
	req := []byte{LongReportID, d.index, featureIndex, functionIndex<<4 | d.receiver.softwareId}
	return d.receiver.Request(append(req, data...))
}

func (d *Device) FeatureRequest(featureId uint16, functionIndex byte, data ...byte) (response []byte, featureIndex byte, err error) {
	var exists bool
	featureIndex, exists = d.GetFeatureIndex(featureId)
	if !exists {
		err = fmt.Errorf("[0x%04X] feature not supported by this device (0x%02X)", featureId, d.index)
		return
	}

	response, err = d.Request(featureIndex, functionIndex, data...)
	return
}
