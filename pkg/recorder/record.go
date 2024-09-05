package recorder

import (
	"sync"
	"time"

	"github.com/prometheus-community/pro-bing"
)

// Record is a record of packets.
type Record struct {
	// packets is a map of sequence numbers to slices of recorded packets.
	packets map[int][]*probing.Packet
	// canStat is a slice of sequence numbers that have enough packets to be able to calculate statistics.
	canStat []int
	// mutex is a mutex to protect the record.
	mutex *sync.RWMutex
}

// NewRecord creates a new record with the given number of probes.
func NewRecord() *Record {
	return &Record{
		packets: make(map[int][]*probing.Packet),
		mutex:   &sync.RWMutex{},
		canStat: make([]int, 0, 100),
	}
}

func (r *Record) AddPacket(p *probing.Packet) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, ok := r.packets[p.Seq]; !ok {
		r.packets[p.Seq] = make([]*probing.Packet, 0, 5)
	}

	r.packets[p.Seq] = append(r.packets[p.Seq], p)

	if len(r.packets[p.Seq]) >= 2 {
		r.canStat = append(r.canStat, p.Seq)
	}
}

func (r *Record) Stats() []*Stat {
	completes := make([][]*probing.Packet, 0, len(r.canStat))

	r.mutex.RLock()
	for _, seq := range r.canStat {
		completes = append(completes, r.packets[seq])
	}
	r.mutex.RUnlock()

	stats := make([]*Stat, 0, len(completes))
	for _, packets := range completes {
		s1 := float64(packets[0].Nbytes)
		rtt1 := packets[0].Rtt.Seconds()
		s2 := float64(packets[1].Nbytes)
		rtt2 := packets[1].Rtt.Seconds()

		// Make sure smallest rtt matches with smallest size

		br := 2 * (s1 - s2) / (rtt1 - rtt2)
		l := (rtt1 - 2*s1/br) / 2

		stats = append(stats, &Stat{
			Bandwidth: br,
			Latency:   time.Duration(l * float64(time.Second)),
			Rtt:       packets[0].Rtt + packets[1].Rtt,
			Seq:       packets[0].Seq,
		})
	}
	return stats
}
