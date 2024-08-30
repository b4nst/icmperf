package recorder

import (
	"sort"
	"time"
)

// Stats contains the statistics of a transfer.
type Stats struct {
	// Latencies is a map latency per nanosecond.
	Latencies map[int64]time.Duration
	// Bandwidths is a map of bandwidth per nanosecond.
	Bandwidths map[int64]float64
	// TotalSent is the total number of packets sent.
	TotalSent int
	// TotalReceived is the total number of packets received.
	TotalReceived int
}

// Bandwidth returns the bandwidth in bytes per second.
// 10% of the best and worst bandwidths are removed before calculating the mean.
func (s *Stats) Bandwidth() float64 {
	bws := make([]float64, 0, len(s.Bandwidths))
	for _, v := range s.Bandwidths {
		bws = append(bws, v)
	}
	sort.Float64s(bws)
	offset := len(bws) / 10
	bws = bws[offset : len(bws)-offset]

	var total float64
	for _, bandwidth := range bws {
		total += bandwidth
	}
	return total / float64(len(bws))
}

// Latency returns the average latency.
// 10% of the best and worst latencies are removed before calculating the mean.
func (s *Stats) Latency() time.Duration {
	latencies := make([]time.Duration, 0, len(s.Latencies))
	for _, v := range s.Latencies {
		latencies = append(latencies, v)
	}
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})
	offset := len(latencies) / 10
	latencies = latencies[offset : len(latencies)-offset]

	var total time.Duration
	for _, latency := range latencies {
		total += latency
	}
	return total / time.Duration(len(latencies))
}

// PacketLoss returns the percentage of packets lost.
func (s *Stats) PacketLoss() float64 {
	return float64(s.TotalSent-s.TotalReceived) / float64(s.TotalSent)
}
