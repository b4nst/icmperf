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

// iqrFilter filters the data using the Interquartile Range method.
func iqrFilter(data []float64) []float64 {
	sort.Float64s(data)
	if len(data) < 4 {
		// Not enough data to filter
		return data
	}

	q1 := data[len(data)/4]
	q3 := data[3*len(data)/4]
	iqr := q3 - q1
	lower := q1 - 1.5*iqr
	upper := q3 + 1.5*iqr

	lowerIndex := 0
	for i, v := range data {
		if v >= lower {
			lowerIndex = i
			break
		}
	}

	upperIndex := len(data)
	for i := len(data) - 1; i >= 0; i-- {
		if data[i] <= upper {
			upperIndex = i
			break
		}
	}

	return data[lowerIndex:upperIndex]
}

// mean calculates the arithmetic mean of a serie.
func mean(data []float64) float64 {
	var total float64
	for _, v := range data {
		total += v
	}
	return total / float64(len(data))
}

// Bandwidth returns the bandwidth in bytes per second.
// The serie is IQR filtered, and mean is calculated.
func (s *Stats) Bandwidth() float64 {
	bws := make([]float64, 0, len(s.Bandwidths))
	for _, v := range s.Bandwidths {
		bws = append(bws, v)
	}
	bws = iqrFilter(bws)

	return mean(bws)
}

// Latency returns the average latency.
// The serie is IQR filtered, and mean is calculated.
func (s *Stats) Latency() time.Duration {
	latencies := make([]float64, 0, len(s.Latencies))
	for _, v := range s.Latencies {
		latencies = append(latencies, v.Seconds())
	}
	latencies = iqrFilter(latencies)

	return time.Duration(mean(latencies) * float64(time.Second))
}

// PacketLoss returns the percentage of packets lost.
func (s *Stats) PacketLoss() float64 {
	return float64(s.TotalSent-s.TotalReceived) / float64(s.TotalSent)
}
