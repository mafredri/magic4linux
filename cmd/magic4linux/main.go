package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bendahl/uinput"

	"github.com/mafredri/magic4linux/m4p"
)

const (
	broadcastPort    = 42830
	subscriptionPort = 42831
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		panic(err)
	}
}

func run(ctx context.Context) error {
	kbd, err := uinput.CreateKeyboard("/dev/uinput", []byte("magic4linux-keyboard"))
	if err != nil {
		return err
	}
	defer kbd.Close()

	tp, err := uinput.CreateTouchPad("/dev/uinput", []byte("magic4linux-touchpad"), 0, 1920, 0, 1080)
	if err != nil {
		return err
	}
	defer tp.Close()

	d, err := m4p.NewDiscoverer(broadcastPort)
	if err != nil {
		return err
	}
	defer d.Close()

	for {
		select {
		case <-ctx.Done():
			return nil

		case dev := <-d.NextDevice():
			err = connect(ctx, dev, kbd, tp)
			if err != nil {
				log.Printf("connect: %v", err)
			}
		}
	}
}

func connect(ctx context.Context, dev m4p.DeviceInfo, kbd uinput.Keyboard, tp uinput.TouchPad) error {
	addr := fmt.Sprintf("%s:%d", dev.IPAddr, dev.Port)
	log.Printf("connect: connecting to: %s", addr)

	client, err := m4p.Dial(ctx, addr)
	if err != nil {
		return err
	}
	defer client.Close()

	for {
		m, err := client.Recv(ctx)
		if err != nil {
			return err
		}

		switch m.Type {
		case m4p.InputMessage:
			log.Printf("connect: got %s: %v", m.Type, m.Input)

			// PoC Kodi keyboard mapping.
			key := m.Input.Parameters.KeyCode
			switch key {
			case m4p.KeyWheelPressed:
				key = uinput.KeyEnter
			case m4p.KeyChannelUp:
				key = uinput.KeyPageup
			case m4p.KeyChannelDown:
				key = uinput.KeyPagedown
			case m4p.KeyLeft:
				key = uinput.KeyLeft
			case m4p.KeyUp:
				key = uinput.KeyUp
			case m4p.KeyRight:
				key = uinput.KeyRight
			case m4p.KeyDown:
				key = uinput.KeyDown
			case m4p.Key0:
				key = uinput.Key0
			case m4p.Key1:
				key = uinput.Key1
			case m4p.Key2:
				key = uinput.Key1
			case m4p.Key3:
				key = uinput.Key1
			case m4p.Key4:
				key = uinput.Key1
			case m4p.Key5:
				key = uinput.Key1
			case m4p.Key6:
				key = uinput.Key1
			case m4p.Key7:
				key = uinput.Key1
			case m4p.Key8:
				key = uinput.Key1
			case m4p.Key9:
				key = uinput.Key1
			case m4p.KeyRed:
				key = uinput.KeyStop
			case m4p.KeyGreen:
				key = uinput.KeyPlaypause
			case m4p.KeyYellow:
				key = uinput.KeyZ
			case m4p.KeyBlue:
				key = uinput.KeyC
			case m4p.KeyBack:
				key = uinput.KeyBackspace
			}

			if m.Input.Parameters.IsDown {
				kbd.KeyDown(key)
			} else {
				kbd.KeyUp(key)
			}
		case m4p.RemoteUpdateMessage:
			// log.Printf("connect: got %s: %s", m.Type, hex.EncodeToString(m.RemoteUpdate.Payload))

			r := bytes.NewReader(m.RemoteUpdate.Payload)
			var returnValue, deviceID uint8
			var coordinate [2]int32
			var gyroscope, acceleration [3]float32
			var quaternion [4]float32
			for _, fn := range []func() error{
				func() error { return binary.Read(r, binary.LittleEndian, &returnValue) },
				func() error { return binary.Read(r, binary.LittleEndian, &deviceID) },
				func() error { return binary.Read(r, binary.LittleEndian, coordinate[:]) },
				func() error { return binary.Read(r, binary.LittleEndian, gyroscope[:]) },
				func() error { return binary.Read(r, binary.LittleEndian, acceleration[:]) },
				func() error { return binary.Read(r, binary.LittleEndian, quaternion[:]) },
			} {
				if err := fn(); err != nil {
					log.Printf("connect: %s decode failed: %v", m.Type, err)
					break
				}
			}

			x := coordinate[0]
			y := coordinate[1]
			fmt.Println("Move mouse", x, y)
			tp.MoveTo(x, y)

			// log.Printf("connect: %d %d %#v %#v %#v %#v", returnValue, deviceID, coordinate, gyroscope, acceleration, quaternion)

		default:
		}
	}
}
