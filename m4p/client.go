package m4p

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

// Client represents an active connection to the magic4pc server.
type Client struct {
	ctx             context.Context
	cancel          context.CancelFunc
	conn            net.Conn
	opts            dialOptions
	serverKeepalive chan struct{}
	recvBuf         chan Message
}

type dialOptions struct {
	updateFrequency int
	filters         []string
}

// DialOption sets options for dial.
type DialOption func(*dialOptions)

// WithUpdateFrequency sets the RemoteUpdate frequency.
func WithUpdateFrequency(f int) func(*dialOptions) {
	return func(o *dialOptions) {
		o.updateFrequency = f
	}
}

// WithFilters specifies the filters used for RemoteUpdate.
func WithFilters(filters ...string) func(*dialOptions) {
	return func(o *dialOptions) {
		o.filters = filters
	}
}

// Dial connects to a magic4pc server running in webOS.
func Dial(ctx context.Context, addr string, opts ...DialOption) (*Client, error) {
	o := dialOptions{
		updateFrequency: 250,
		filters:         DefaultFilters,
	}
	for _, opt := range opts {
		opt(&o)
	}

	d := &net.Dialer{
		Timeout: 5 * time.Second,
	}
	conn, err := d.DialContext(ctx, "udp4", addr)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	c := &Client{
		ctx:             ctx,
		cancel:          cancel,
		conn:            conn,
		opts:            o,
		serverKeepalive: make(chan struct{}, 1),
		recvBuf:         make(chan Message, 10), // Buffer up to 10 messages after which we block.
	}

	// Register our client with the server.
	m := NewMessage(SubSensorMessage)
	m.Register = &Register{
		UpdateFrequency: c.opts.updateFrequency,
		Filter:          c.opts.filters,
	}

	err = c.Send(m)
	if err != nil {
		c.Close()
		return nil, fmt.Errorf("register failed: %w", err)
	}

	// Tell the server that we're alive and well.
	go c.keepalive()
	go c.recv()

	return c, nil
}

func (c *Client) recv() {
	var buf [1024]byte
recvLoop:
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		n, err := c.conn.Read(buf[:])
		if err != nil {
			log.Printf("m4p: Client: recv: read udp packet failed: %v", err)
			continue
		}

		m, err := decode(buf[:n])
		if err != nil {
			log.Printf("m4p: Client: recv: decode failed: %v", err)
			continue
		}

		switch m.Type {
		case KeepAliveMessage:
			// log.Printf("m4p: Client: recv: got %s", m.Type)

			// Trigger server keepalive, non-blocking (chan is buffered).
			select {
			case c.serverKeepalive <- struct{}{}:
			default:
			}

			goto recvLoop

		case InputMessage:
			log.Printf("m4p: Client: recv: got %s: %v", m.Type, m.Input)

		case RemoteUpdateMessage:
			log.Printf("m4p: Client: recv: got %s: %s", m.Type, hex.EncodeToString(m.RemoteUpdate.Payload))

		default:
			log.Printf("m4p: Client: recv: unknown message: %s", m.Type)
		}

		select {
		case c.recvBuf <- m:
		default:
			log.Printf("m4p: Client: recv: buffer full, discarding message: %s", m.Type)
		}
	}
}

func (c *Client) keepalive() {
	defer c.Close()

	serverDeadline := time.After(keepaliveTimeout)
	clientKeepalive := time.After(clientKeepaliveInterval)

	for {
		select {
		case <-c.ctx.Done():
			return

		case <-c.serverKeepalive:
			serverDeadline = time.After(keepaliveTimeout)

		case <-serverDeadline:
			log.Printf("m4p: Client: keepalive: server keepalive deadline reached, disconnecting...")
			return

		case <-clientKeepalive:
			_, err := c.conn.Write([]byte("{}"))
			if err != nil {
				log.Printf("m4p: Client: keepalive: send client keepalive failed, disconnecting...")
				return
			}
			clientKeepalive = time.After(clientKeepaliveInterval)
		}
	}
}

// Send a message to the magic4pc server.
func (c *Client) Send(m Message) error {
	b, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("json encode message failed: %w", err)
	}
	if _, err = c.conn.Write(b); err != nil {
		return fmt.Errorf("write message failed: %w", err)
	}
	return nil
}

// Recv messages from the magic4pc server. Keepalives are handled
// transparently by the client and are not observable.
func (c *Client) Recv(ctx context.Context) (Message, error) {
	select {
	case <-ctx.Done():
		return Message{}, ctx.Err()
	case <-c.ctx.Done():
		return Message{}, c.ctx.Err()
	case m := <-c.recvBuf:
		return m, nil
	}
}

// Close the client and connection.
func (c *Client) Close() error {
	c.cancel()
	return c.conn.Close()
}
