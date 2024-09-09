package pinger

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-netroute"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type Pinger struct {
	// Size is the size of the data probe to send. If 0, computed using interface MTU.
	Size int
	// Timeout is the evaluation timeout.
	Timeout time.Duration

	// conn is the ICMP packet connection.
	conn *icmp.PacketConn
	// id is the ICMP packet identifier.
	id int
	// seq is the ICMP packet sequence number.
	seq atomic.Int32
	// sent is the number of sent packets.
	sent atomic.Int32
	// received is the number of received packets.
	received atomic.Int32
	// Time of the first sent packet, in nanoseconds since the Unix epoch
	firstSent atomic.Int64
	// Time of the last received packet, in nanoseconds since the Unix epoch
	lastReceived atomic.Int64
	readbuf      []byte
	data         []byte
}

func NewPinger(size int, timeout time.Duration) *Pinger {
	return &Pinger{
		Size:    size,
		Timeout: timeout,
		id:      os.Getpid() & 0xffff,
	}
}

func (p *Pinger) connect(target *net.IPAddr) error {
	r, err := netroute.New()
	if err != nil {
		return fmt.Errorf("get routing table: %v", err)
	}
	iface, _, src, err := r.Route(target.IP)
	if err != nil {
		return fmt.Errorf("get route: %v", err)
	}

	// Create a new ICMP packet connection
	conn, err := icmp.ListenPacket("ip4:icmp", src.String())
	if err != nil {
		return err
	}
	p.conn = conn

	p.reset(iface.MTU)

	return nil
}

func (p *Pinger) reset(mtu int) {
	p.readbuf = make([]byte, mtu)
	size := p.Size
	if size <= 0 {
		size = mtu - 8 - 20
	}
	p.data = make([]byte, size)

	p.sent.Store(0)
	p.received.Store(0)
	p.firstSent.Store(0)
	p.lastReceived.Store(0)
}

func (p *Pinger) Close() error {
	if p.conn == nil {
		return nil
	}

	return p.conn.Close()
}

func (p *Pinger) Eval(target *net.IPAddr) (*Statistics, error) {
	if p.conn != nil {
		p.Close()
	}
	if err := p.connect(target); err != nil {
		return nil, err
	}
	defer p.Close()

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(p.Timeout))
	defer cancel()

	go p.receive(ctx)      // Start receiving packets
	go p.echo(ctx, target) // Start sending packets

	<-ctx.Done()

	return p.Stats(), nil
}

func (p *Pinger) Stats() *Statistics {
	return &Statistics{
		Sent:       int(p.sent.Load()),
		Received:   int(p.received.Load()),
		Duration:   time.Duration(p.lastReceived.Load() - p.firstSent.Load()),
		PacketSize: len(p.data) + 8,
	}
}

func (p *Pinger) receive(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, _, err := p.conn.ReadFrom(p.readbuf)
			if err != nil {
				if ctx.Err() == nil {
					fmt.Printf("read icmp message: %v\n", err)
				}
				continue
			}
			recpt := time.Now()
			msg, err := icmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), p.readbuf[:n])
			if err != nil {
				fmt.Printf("parse icmp message: %v\n", err)
				continue
			}
			body, ok := msg.Body.(*icmp.Echo)
			if !ok {
				fmt.Println("invalid icmp message body")
				continue
			}
			if body.ID != p.id {
				fmt.Println("invalid icmp message id")
				continue
			}
			p.received.Add(1)
			p.lastReceived.Store(recpt.UnixNano())
		}
	}
}

// Ping sends an ICMP echo request to the target
// It is safe to call Ping concurrently
func (p *Pinger) echo(ctx context.Context, target *net.IPAddr) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Create a icmp message
			msg := icmp.Message{
				Type: ipv4.ICMPTypeEcho,
				Code: 0,
				Body: &icmp.Echo{
					ID:   p.id,
					Seq:  int(p.seq.Add(1)),
					Data: p.data,
				},
			}
			b, err := msg.Marshal(nil)
			if err != nil {
				continue
			}

			// Send the message
			_, err = p.conn.WriteTo(b, target)
			if err != nil {
				continue
			}

			if p.firstSent.Load() == 0 {
				p.firstSent.Store(time.Now().UnixNano())
			}

			p.sent.Add(1)
		}
	}
}
