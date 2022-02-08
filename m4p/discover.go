package m4p

import (
	"errors"
	"log"
	"net"
)

// Discoverer magic4pc servers.
type Discoverer struct {
	ln     *net.UDPConn
	device chan DeviceInfo
}

// NewDiscover returns a new Discoverer that listens on the broadcast port.
func NewDiscoverer(broadcastPort int) (*Discoverer, error) {
	addr := net.UDPAddr{
		Port: broadcastPort,
		IP:   net.ParseIP("0.0.0.0"),
	}

	ln, err := net.ListenUDP("udp", &addr)
	if err != nil {
		panic(err)
	}

	d := &Discoverer{
		ln:     ln,
		device: make(chan DeviceInfo), // Unbuffered, discard when nobody is listening.
	}
	go d.discover()

	return d, nil
}

func (d *Discoverer) discover() {
	var buf [1024]byte
	for {
		n, addr, err := d.ln.ReadFromUDP(buf[:])
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				log.Printf("m4p: Discoverer: discover: connection closed: %v", err)
				return
			}
			log.Printf("m4p: Discoverer: discover: read udp packet failed: %v", err)
			continue
		}

		m, err := decode(buf[:n])
		if err != nil {
			log.Printf("m4p: Discoverer: discover: decode failed: %v", err)
		}

		switch m.Type {
		case Magic4PCAdMessage:
			dev := m.DeviceInfo
			dev.IPAddr = addr.IP.String()
			log.Printf("m4p: Discoverer: discover: found device: %#v", dev)

			select {
			case d.device <- *dev:
			default:
			}

		default:
			log.Printf("m4p: Discoverer: discover: unknown message: %s", m.Type)
		}
	}
}

// NextDevice returns a newly discovered magic4pc server.
func (d *Discoverer) NextDevice() <-chan DeviceInfo {
	return d.device
}

// Close the Discoverer and stop listening for broadcasts.
func (d *Discoverer) Close() error {
	return d.ln.Close()
}
