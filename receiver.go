package bolt

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/sstallion/go-hid"
)

const (
	vidLogitech   = 0x046d
	pidBolt       = 0xc548
	usagePageBolt = 0xFF00
	usageBolt     = 0x0001
)

type HandlerFunc func([]byte)

type Packet struct {
	Data []byte
}

type Receiver struct {
	Path    string
	Name    string
	handler HandlerFunc

	// private
	dev        *hid.Device
	mu         sync.Mutex
	waiters    map[string]chan Packet
	softwareId byte // software IDs 0x01..0x0F, default is 0x07
	quit       chan struct{}
	opened     bool
}

func (c *Receiver) SetHandler(handler HandlerFunc) error {
	if handler != nil {
		c.handler = handler
		return nil
	}
	return fmt.Errorf("invalid handler set")
}

func (c *Receiver) SetSoftwareId(softwareId byte) error {
	if softwareId >= 0x01 && softwareId <= 0x0F {
		c.softwareId = softwareId
		return nil
	}
	return fmt.Errorf("software id out of range: 0x%02X", softwareId)
}

func (c *Receiver) Open() error {
	slog.Debug("Opening Bolt Receiver")
	var err error

	c.mu.Lock()
	opened := c.opened
	if !opened {
		c.opened = true
		c.quit = make(chan struct{})
	}
	c.mu.Unlock()

	if opened {
		return fmt.Errorf("receiver already opened")
	}

	if c.dev, err = hid.OpenPath(c.Path); err != nil {
		c.mu.Lock()
		c.opened = false
		close(c.quit)
		c.mu.Unlock()
		return err
	}

	go c.listen()

	return nil
}

func (c *Receiver) Close() error {
	slog.Debug("Closing Bolt Receiver")

	c.mu.Lock()
	opened := c.opened
	if opened {
		c.opened = false
		close(c.quit)
	}
	c.mu.Unlock()

	if !opened {
		return fmt.Errorf("receiver already closed")
	}

	return c.dev.Close()
}

func (c *Receiver) listen() {
	buf := make([]byte, 64)

	slog.Debug("Listening for packets...")

	for {
		select {
		case <-c.quit:
			return
		default:
		}

		n, err := c.dev.ReadWithTimeout(buf, 250*time.Millisecond)
		if err != nil {
			//if !errors.Is(err, hid.ErrTimeout) {
			//	slog.Error("ошибка чтения", "err", err)
			//}
			continue
		}

		if n == 0 {
			// timeout
			continue
		}

		pkt := append([]byte(nil), buf[:n]...)

		key := packetKey(pkt)

		slog.Debug("R", "<-", debugValue(pkt), "key", key)

		c.mu.Lock()
		ch := c.waiters[key]
		c.mu.Unlock()

		if ch != nil {
			select {
			case ch <- Packet{Data: pkt}:
			default:
			}
		} else if c.handler != nil {
			c.handler(pkt)
		}
	}
}

func All() []*Receiver {
	b := make([]*Receiver, 0)
	_ = hid.Enumerate(vidLogitech, pidBolt, func(di *hid.DeviceInfo) error {
		if di.VendorID == vidLogitech && di.ProductID == pidBolt && di.UsagePage == usagePageBolt && di.Usage == usageBolt {
			b = append(b, &Receiver{
				Path:       di.Path,
				Name:       fmt.Sprintf("%s %s", di.MfrStr, di.ProductStr),
				softwareId: 0x07,
			})
		}
		return nil
	})
	return b
}

func First() *Receiver {
	all := All()
	if len(all) > 0 {
		return all[0]
	}
	return nil
}

func packetKey(pkt []byte) string {
	if len(pkt) < 4 {
		return ""
	}

	device := pkt[1]

	// HID++ error packet
	if pkt[2] == 0x8F {
		feature := pkt[3]
		function := pkt[4] & 0xF0

		return fmt.Sprintf("%02x:%02x:%02x", device, feature, function)
	}

	feature := pkt[2]
	function := pkt[3] & 0xF0

	return fmt.Sprintf("%02x:%02x:%02x", device, feature, function)
}

func debugValue(d []byte) string {
	return fmt.Sprintf("[%02X %02X %X %X]", d[0], d[1], d[2:4], d[4:])
}
