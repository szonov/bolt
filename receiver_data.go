package bolt

func (c *Receiver) GetSerialNumber() (string, error) {
	res, err := c.Request([]byte{ShortReportID, 0xFF, 0x83, 0xFB})
	if err != nil {
		return "", err
	}

	return string(res[4:]), nil
}
