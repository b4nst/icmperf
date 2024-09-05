package recorder

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus-community/pro-bing"
	"github.com/stretchr/testify/assert"
)

func TestNewRecord(t *testing.T) {
	t.Parallel()
	r := NewRecord()
	assert.NotNil(t, r)
	assert.IsType(t, &Record{}, r)
}

func TestRecord_AddPacket(t *testing.T) {
	t.Parallel()

	t.Run("first packet", func(t *testing.T) {
		t.Parallel()
		r := NewRecord()
		p := &probing.Packet{Seq: 1}
		r.AddPacket(p)
		assert.Len(t, r.packets, 1)
		assert.Len(t, r.packets[1], 1)
		assert.Equal(t, p, r.packets[1][0])
	})

	t.Run("two packets, same sequence", func(t *testing.T) {
		t.Parallel()
		r := NewRecord()
		p1 := &probing.Packet{Seq: 1}
		p2 := &probing.Packet{Seq: 1}
		r.AddPacket(p1)
		r.AddPacket(p2)
		assert.Len(t, r.packets, 1)
		assert.Len(t, r.packets[1], 2)
		assert.Equal(t, p1, r.packets[1][0])
		assert.Equal(t, p2, r.packets[1][1])
		assert.ElementsMatch(t, r.canStat, []int{1})
	})

	t.Run("two packets, different sequence", func(t *testing.T) {
		t.Parallel()
		r := NewRecord()
		p1 := &probing.Packet{Seq: 1}
		p2 := &probing.Packet{Seq: 2}
		r.AddPacket(p1)
		r.AddPacket(p2)
		assert.Len(t, r.packets, 2)
		assert.Len(t, r.packets[1], 1)
		assert.Len(t, r.packets[2], 1)
		assert.Equal(t, p1, r.packets[1][0])
		assert.Equal(t, p2, r.packets[2][0])
		assert.Len(t, r.canStat, 0)
	})

	t.Run("concurrent", func(t *testing.T) {
		t.Parallel()
		r := NewRecord()
		var seq atomic.Int32
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				p := &probing.Packet{Seq: int(seq.Add(1))}
				r.AddPacket(p)
			}()
		}
		wg.Wait()
		assert.Len(t, r.packets, 100)
	})
}

func TestRecord_Stats(t *testing.T) {
	t.Parallel()

	t.Run("no stats", func(t *testing.T) {
		t.Parallel()
		r := NewRecord()
		r.packets = map[int][]*probing.Packet{
			1: []*probing.Packet{&probing.Packet{Seq: 1}},
		}
		stats := r.Stats()
		assert.Len(t, stats, 0)
	})

	t.Run("2 packets stat", func(t *testing.T) {
		latency := 12 * time.Millisecond
		bitrate := 1000.0 // bytes per second

		t.Parallel()
		r := NewRecord()
		p1 := &probing.Packet{Seq: 1, Nbytes: 24, Rtt: 2*latency + time.Duration(2*24/bitrate*float64(time.Second))}
		p2 := &probing.Packet{Seq: 1, Nbytes: 124, Rtt: 2*latency + time.Duration(2*124/bitrate*float64(time.Second))}
		r.AddPacket(p1)
		r.AddPacket(p2)
		stats := r.Stats()
		assert.Len(t, stats, 1)
		assert.Equal(t, 1, stats[0].Seq)
		assert.InEpsilon(t, latency, stats[0].Latency, 0.0001)
		assert.Equal(t, p1.Rtt+p2.Rtt, stats[0].Rtt)
		assert.Equal(t, bitrate, stats[0].Bandwidth)
	})
}
