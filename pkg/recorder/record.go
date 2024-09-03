package recorder

import (
	"sync"
	"time"

	"github.com/emirpasic/gods/maps/treemap"
	"github.com/emirpasic/gods/utils"
)

type TransferMetric struct {
	Received time.Time
	Sent     time.Time
	Bytes    int
}

type Record struct {
	inflight map[uint64]*TransferMetric

	latencies *treemap.Map
	tms       []*TransferMetric

	inlock   sync.Mutex
	recvlock sync.Mutex
}

func NewRecord() *Record {
	return &Record{
		inflight:  make(map[uint64]*TransferMetric),
		latencies: treemap.NewWith(utils.Int64Comparator),
		tms:       []*TransferMetric{},

		inlock:   sync.Mutex{},
		recvlock: sync.Mutex{},
	}
}

func (s *Record) PacketSent(id uint64, size int, sent time.Time) {
	s.inlock.Lock()
	defer s.inlock.Unlock()

	s.inflight[id] = &TransferMetric{
		Bytes: size,
		Sent:  sent,
	}
}

func (s *Record) PacketReceived(id uint64, size int, received time.Time) {
	// Remove the packet from the inflight map
	s.inlock.Lock()
	tm, ok := s.inflight[id]
	delete(s.inflight, id)
	s.inlock.Unlock()
	if !ok {
		return
	}

	s.recvlock.Lock()
	defer s.recvlock.Unlock()
	if tm.Bytes == 0 {
		// Record Latency ping
		s.latencies.Put(tm.Sent.UnixNano(), received.Sub(tm.Sent))
	} else {
		// Record bandwidth ping
		tm.Received = received
		tm.Bytes += size
		s.tms = append(s.tms, tm)
	}
}

func (s *Record) Stats() *Stats {
	s.inlock.Lock()
	defer s.inlock.Unlock()
	s.recvlock.Lock()
	defer s.recvlock.Unlock()

	totalLost := len(s.inflight)
	totalPackets := totalLost + len(s.tms) + s.latencies.Size()

	// Calculate bandwidths
	bws := map[int64]float64{}
	for _, tm := range s.tms {
		// Find the closest latency to the current bandwidth packet
		k, v := s.latencies.Floor(tm.Sent.UnixNano())
		if v == nil {
			k, v = s.latencies.Ceiling(tm.Sent.UnixNano())
			if k == nil {
				// No latencies recorded, skip this bandwidth packet
				continue
			}
		}
		latency := v.(time.Duration)
		bw := float64(tm.Bytes) / (tm.Received.Sub(tm.Sent) - latency).Seconds()
		// Record the bandwidth if it is not NaN
		if bw == bw {
			bws[tm.Sent.UnixNano()] = bw
		}
	}

	latencies := map[int64]time.Duration{}
	s.latencies.Each(func(key interface{}, value interface{}) {
		latencies[key.(int64)] = value.(time.Duration)
	})

	return &Stats{
		Latencies:     latencies,
		Bandwidths:    bws,
		TotalSent:     totalPackets,
		TotalReceived: totalPackets - totalLost,
	}
}
