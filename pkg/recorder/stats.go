package recorder

import (
	"fmt"
	"time"

	"github.com/montanaflynn/stats"
)

// Stats contains the statistics of a transfer.
type Stat struct {
	// Latency for the transfer.
	Latency time.Duration
	// Total time for the transfer.
	Rtt time.Duration
	// Bandwidth for the transfer.
	Bandwidth float64
	// The sequence number of the transfer.
	Seq int
}

func ProcessStats(stats []*Stat) (*Stat, error) {
	sanitized := Sanitize(stats)
	fmt.Println(sanitized)
	// Apply IQR filter
	sanitized, err := IQRFilter(sanitized)
	if err != nil {
		return nil, err
	}
	fmt.Println(sanitized)
	// Return average
	return Average(sanitized), nil
}

func Sanitize(stats []*Stat) []*Stat {
	sanitized := make([]*Stat, 0, len(stats))
	for _, stat := range stats {
		if stat.Bandwidth > 0 {
			sanitized = append(sanitized, stat)
		}
	}
	return sanitized
}

func IQRFilter(s []*Stat) ([]*Stat, error) {
	bandwidths := make(stats.Float64Data, 0, len(s))
	for _, stat := range s {
		bandwidths = append(bandwidths, stat.Bandwidth)
	}

	quartiles, err := bandwidths.Quartiles()
	if err != nil {
		return nil, err
	}
	iqr := quartiles.Q3 - quartiles.Q1

	lb := quartiles.Q1 - 1.5*iqr
	ub := quartiles.Q3 + 1.5*iqr
	filtered := make([]*Stat, 0, len(s))
	for _, stat := range s {
		if stat.Bandwidth >= lb && stat.Bandwidth <= ub {
			filtered = append(filtered, stat)
		}
	}

	return filtered, nil
}

func Average(stats []*Stat) *Stat {
	total := Stat{Seq: -1}
	for _, stat := range stats {
		total.Latency += stat.Latency
		total.Rtt += stat.Rtt
		total.Bandwidth += stat.Bandwidth
	}
	total.Latency /= time.Duration(len(stats))
	total.Rtt /= time.Duration(len(stats))
	total.Bandwidth /= float64(len(stats))
	return &total
}
