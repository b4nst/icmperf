package pinger

import (
	"context"
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

	recvCb func(*icmp.Message, time.Time) error
}

func NewPinger(addr string, mtu int, timeout time.Duration) *Pinger {
	return &Pinger{
		listenaddr: addr,
		timeout:    timeout,
		MTU:        mtu,
		seq:        atomic.Uint64{},
		id:         os.Getpid() & 0xffff,
	}
}

func (p *Pinger) Start(ctx context.Context) error {
	conn, err := icmp.ListenPacket("udp4", p.listenaddr)
	if err != nil {
		return err
	}
	p.conn = conn
	go p.listen(ctx)
	return nil
}

func (p *Pinger) OnRecv(cb func(*icmp.Message, time.Time) error) {
	p.recvCb = cb
}

func (p *Pinger) listen(ctx context.Context) {
	buf := make([]byte, p.MTU)
	for {
		if ctx.Err() != nil {
			return
		}
		err := p.conn.SetReadDeadline(time.Now().Add(p.timeout))
		if err != nil {
			// TODO: log error
			continue
		}
		n, _, err := p.conn.ReadFrom(buf)
		if err != nil {
			// TODO: log error
			continue
		}
		received := time.Now()
		parsed, err := icmp.ParseMessage(1, buf[:n])
		if err != nil {
			// TODO: log error
			continue
		}
		if err := p.recvCb(parsed, received); err != nil {
			// TODO: log error
			continue
		}
	}
}

func (p *Pinger) Send(peer *net.UDPAddr, payload []byte) (uint64, time.Time, error) {
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
