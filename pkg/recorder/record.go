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
		var bandwidth, latency float64
		var rtt time.Duration

		rtt += packets[0].Rtt
		for i := 1; i < len(packets); i++ {
			rtt += packets[i].Rtt
			localbw := float64(packets[i].Nbytes-packets[i-1].Nbytes) / (packets[i].Rtt - packets[i-1].Rtt).Seconds()
			locallat := packets[i].Rtt.Seconds() - (float64(packets[i].Nbytes) / localbw)

			bandwidth += localbw
			latency += locallat
		}

		stats = append(stats, &Stat{
			Bandwidth: bandwidth / float64(len(packets)-1),
			Latency:   time.Duration(latency / float64(len(packets)-1) * float64(time.Second)),
			Rtt:       rtt,
			Seq:       packets[0].Seq,
		})
	}
	return stats
}
