package bolt

import "fmt"

type Hidpp1Error byte

func (e Hidpp1Error) Error() string {
	switch e {
	case 0x0:
		return "1.0 [0x00] No error / undefined"
	case 0x1:
		return "1.0 [0x01] Invalid SubID / command"
	case 0x2:
		return "1.0 [0x02] Invalid address"
	case 0x3:
		return "1.0 [0x03] Invalid value"
	case 0x4:
		return "1.0 [0x04] Connection request failed (Receiver)"
	case 0x5:
		return "1.0 [0x05] Too many devices connected (Receiver)"
	case 0x6:
		return "1.0 [0x06] Already exists (Receiver)"
	case 0x7:
		return "1.0 [0x07] Busy (Receiver)"
	case 0x8:
		return "1.0 [0x08] Unknown device (Receiver)"
	case 0x9:
		return "1.0 [0x09] Resource error (Receiver)"
	case 0xA:
		return "1.0 [0x0A] Request not valid in current context"
	case 0xB:
		return "1.0 [0x0B] Request parameter has unsupported value"
	case 0xC:
		return "1.0 [0x0C] the PIN code entered on the device was wrong"
	}
	return fmt.Sprintf("1.0 [0x%02X] unknown error code", e)
}

type Hidpp2Error byte

func (e Hidpp2Error) Error() string {
	switch e {
	case 0:
		return "2.0 [0x00] no error"
	case 1:
		return "2.0 [0x01] unknown"
	case 2:
		return "2.0 [0x02] invalid argument"
	case 3:
		return "2.0 [0x03] out of range"
	case 4:
		return "2.0 [0x04] hardware error"
	case 5:
		return "2.0 [0x05] logitech internal"
	case 6:
		return "2.0 [0x06] invalid feature index"
	case 7:
		return "2.0 [0x07] invalid function"
	case 8:
		return "2.0 [0x08] busy"
	case 9:
		return "2.0 [0x09] unsupported"
	}
	return fmt.Sprintf("2.0 [0x%02X] unknown error code", e)
}
