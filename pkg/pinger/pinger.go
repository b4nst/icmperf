package pinger

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync/atomic"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type Pinger struct {
	conn       *icmp.PacketConn
	listenaddr string
	timeout    time.Duration
	MTU        int
	seq        atomic.Uint64
	id         int

	recvbuf []byte
}

func NewPinger(addr string, mtu int, timeout time.Duration) *Pinger {
	return &Pinger{
		listenaddr: addr,
		timeout:    timeout,
		MTU:        mtu,
		seq:        atomic.Uint64{},
		id:         os.Getpid() & 0xffff,

		recvbuf: make([]byte, mtu),
	}
}

func (p *Pinger) Start(ctx context.Context) error {
	conn, err := icmp.ListenPacket("udp4", p.listenaddr)
	if err != nil {
		return err
	}
	p.conn = conn
	return nil
}

// Reply waits for an ICMP reply and fills the message with the response.
func (p *Pinger) Reply(message *icmp.Message) (time.Time, error) {
	err := p.conn.SetReadDeadline(time.Now().Add(p.timeout))
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to set read deadline: %w", err)
	}
	n, _, err := p.conn.ReadFrom(p.recvbuf)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to read from connection: %w", err)
	}
	received := time.Now()
	parsed, err := icmp.ParseMessage(1, p.recvbuf[:n])
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse message: %w", err)
	}
	*message = *parsed

	return received, nil
}

func (p *Pinger) Echo(peer *net.UDPAddr, payload []byte) (uint64, time.Time, error) {
	body := &icmp.Echo{
		ID:   p.id,
		Seq:  int(p.seq.Add(1)),
		Data: payload,
	}
	id := uint64(body.ID)<<32 | uint64(body.Seq)

	message := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: body,
	}
	messageBytes, err := message.Marshal(nil)
	if err != nil {
		return id, time.Now(), err
	}
	_, err = p.conn.WriteTo(messageBytes, peer)
	return id, time.Now(), err
}

func (p *Pinger) Close() error {
	return p.conn.Close()
}
