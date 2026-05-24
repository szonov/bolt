package bolt

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/sstallion/go-hid"
)

const (
	vidLogitech        = 0x046d
	pidBolt            = 0xc548
	usagePageBolt      = 0xFF00
	usageBolt          = 0x0001
	ShortReportID byte = 0x10
	LongReportID  byte = 0x11
)

var (
	ReadPacketTimeout   = 250 * time.Millisecond
	ResponseWaitTimeout = 2 * time.Second
)

type HandlerFunc func(*Receiver, []byte)

type Packet struct {
	Data  []byte
	Error error
}

// packetKey is a compact binary key for the waiters map.
// Layout: [device:1][feature:1][function:1]
type packetKey struct {
	device   byte
	feature  byte
	function byte
}

type Receiver struct {
	Path    string
	Name    string
	handler HandlerFunc

	// private
	dev        *hid.Device
	mu         sync.Mutex
	waiters    map[packetKey]chan Packet
	softwareId byte // software IDs 0x01..0x0F, default is 0x07
	quit       chan struct{}
	opened     bool
	closeOnce  sync.Once
}

// SetHandler устанавливает обработчик входящих пакетов.
// Обработчик вызывается для любого пакета, не попавшего в канал waiters.
func (c *Receiver) SetHandler(handler HandlerFunc) error {
	c.handler = handler
	return nil
}

// SetSoftwareId устанавливает Software ID устройства.
// Допустимые значения: 0x01..0x0F (по умолчанию 0x07).
// Используется для переключения между различными функциями Bolt.
func (c *Receiver) SetSoftwareId(softwareId byte) error {
	if softwareId >= 0x01 && softwareId <= 0x0F {
		c.softwareId = softwareId
		return nil
	}
	return fmt.Errorf("software id out of range: 0x%02X", softwareId)
}

// Open открывает Bolt device для чтения входящих пакетов.
// Поднимает фоновый goroutine listen() и начинает приём данных по HID.
func (c *Receiver) Open() error {
	slog.Debug("Opening Bolt Receiver")

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.opened {
		return fmt.Errorf("receiver already opened")
	}

	dev, err := hid.OpenPath(c.Path)
	if err != nil {
		return fmt.Errorf("failed to open device: %w", err)
	}

	c.dev = dev
	c.opened = true
	c.quit = make(chan struct{})
	c.waiters = make(map[packetKey]chan Packet)

	go c.listen()

	return nil
}

// close performs the actual cleanup, safe for concurrent/repeated calls.
func (c *Receiver) close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.opened {
		return nil
	}

	c.opened = false
	close(c.quit)

	return c.dev.Close()
}

// Close закрывает Bolt device и освобождает ресурсы.
// Вызов безопасен при многократном выполнении.
func (c *Receiver) Close() (err error) {
	slog.Debug("Closing Bolt Receiver")

	c.closeOnce.Do(func() {
		err = c.close()
	})

	return err
}

// listen запускает фоновый цикл чтения входящих HID пакетов от Bolt устройства.
func (c *Receiver) listen() {
	defer func() {
		c.closeOnce.Do(func() {
			c.close()
		})
	}()

	buf := make([]byte, 64)

	for {
		// non-blocking quit check — prevents hanging Close() under event flood
		select {
		case <-c.quit:
			return
		default:
		}

		n, err := c.dev.ReadWithTimeout(buf, ReadPacketTimeout)

		if err != nil {
			if errors.Is(err, hid.ErrTimeout) {
				continue
			}
			select {
			case <-c.quit:
			default:
				slog.Error("Bolt read error", "err", err)
			}
			return
		}

		if n == 0 {
			continue
		}

		pkt := make([]byte, n)
		copy(pkt, buf)

		key := makePacketKey(pkt)
		// формировать пакет здесь

		// Логирование только если включено debug, чтобы не тратить ресурсы на fmt
		if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
			slog.Debug("R", "<-", debugValue(pkt))
		}

		c.mu.Lock()
		ch := c.waiters[key]
		c.mu.Unlock()

		if ch != nil {
			select {
			case ch <- Packet{Data: pkt}:
			default:
			}
		} else if c.handler != nil {
			c.handler(c, pkt)
		}
	}
}

func (c *Receiver) Request(req []byte, noReplay ...bool) ([]byte, error) {

	if len(req) < 4 {
		return nil, errors.New("invalid request")
	}

	shouldWait := len(noReplay) == 0 || noReplay[0] == false

	var needLen int
	switch req[0] {
	case ShortReportID:
		needLen = 7
	case LongReportID:
		needLen = 20
	default:
		// nothing
	}

	if needLen > 0 && len(req) < needLen {
		tmp := make([]byte, needLen)
		copy(tmp, req)
		req = tmp
	}

	// Логирование только если включено debug, чтобы не тратить ресурсы на fmt
	if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		slog.Debug("W", "->", debugValue(req))
	}

	if !shouldWait {
		_, err := c.dev.Write(req)
		return nil, err
	}

	key := makePacketKey(req)

	// REQ:
	// byte 0  report id
	// byte 1  device index
	// byte 2  feature index
	// byte 3  function id + software id
	// byte 4... payload

	ch := make(chan Packet, 1)

	c.mu.Lock()
	c.waiters[key] = ch
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.waiters, key)
		c.mu.Unlock()
	}()

	_, err := c.dev.Write(req)
	if err != nil {
		return nil, err
	}

	select {
	case pkt := <-ch:
		if pkt.Data[2] == 0x8F { // hid++ 1.0
			return pkt.Data, Hidpp1Error(pkt.Data[5])
		}
		if pkt.Data[2] == 0xFF { // hid++ 2.0
			return pkt.Data, Hidpp2Error(pkt.Data[5])
		}
		return pkt.Data, nil
	case <-c.quit:
		return nil, errors.New("device closed")
	case <-time.After(ResponseWaitTimeout):
		return nil, errors.New("timeout")
	}
}

func (c *Receiver) NewDevice(index byte) *Device {
	return NewDevice(c, index)
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

// makePacketKey извлекает ключ пакета из бинарных данных HID++.
// Ключ содержит: device ID, feature и function (верхние биты).
// Для пакетов ошибок (0x8F) используется специализированная логика парсинга.
func makePacketKey(pkt []byte) packetKey {
	if len(pkt) < 5 {
		return packetKey{}
	}

	device := pkt[1]

	var feature, function byte
	if pkt[2] == 0x8F || pkt[2] == 0xFF {
		// HID++ error packet
		feature = pkt[3]
		function = pkt[4] & 0xF0
	} else {
		feature = pkt[2]
		function = pkt[3] & 0xF0
	}

	return packetKey{device, feature, function}
}

// debugValue формирует человеково-понятную строку отладочного вывода.
// Выделяет верхние биты function, чтобы упростить чтение логов.
func debugValue(d []byte) string {
	if len(d) < 4 {
		return fmt.Sprintf("[%X]", d)
	}
	return fmt.Sprintf("[%02X %02X %X %X]", d[0], d[1], d[2:4], d[4:])
}
