package session

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/montanaflynn/stats"
	"github.com/prometheus-community/pro-bing"
)

type Statistics struct {
	Pingers []*StatWithSize
	Bitrate float64
	Latency time.Duration
}

func NewStatistics(pingers []*probing.Pinger) (*Statistics, error) {
	pstats := make([]*StatWithSize, 0, len(pingers))

	stats0 := pingers[0].Statistics()
	stats1 := pingers[1].Statistics()

	for _, pinger := range pingers {
		stats := pinger.Statistics()
		pstats = append(pstats, &StatWithSize{
			Statistics: stats,
			Bytes:      pinger.Size,
		})
	}

	bitrate := float64(pingers[0].Size-pingers[1].Size) / (stats0.MinRtt - stats1.MinRtt).Seconds()
	latency := (stats0.MinRtt.Seconds() - float64(pingers[0].Size)/bitrate) / 2

	return &Statistics{
		Pingers: pstats,
		Bitrate: bitrate,
		Latency: time.Duration(latency * float64(time.Second)),
	}, nil
}

func (s *Statistics) String() string {
	sb := strings.Builder{}
	for _, pinger := range s.Pingers {
		sb.WriteString(fmt.Sprintf("Sent %d packets of %d bytes\n", pinger.PacketsSent, pinger.Bytes))
		sb.WriteString(fmt.Sprintf("Rtts: %v\n", pinger.Rtts))
	}
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Estimate bitrate: %s/sec latency: %s\n", humanize.Bytes(uint64(s.Bitrate)), s.Latency))
	return sb.String()
}

func sanitize(rtts []time.Duration) (float64, error) {
	r := make(stats.Float64Data, 0, len(rtts))
	for _, rtt := range rtts {
		r = append(r, rtt.Seconds())
	}

	filtered, err := MADFilter(r)
	if err != nil {
		return 0, err
	}

	return stats.Mean(filtered)
}

func MADFilter(data stats.Float64Data) (stats.Float64Data, error) {
	mad, err := stats.MedianAbsoluteDeviation(data)
	if err != nil {
		return nil, err
	}
	if mad == 0 {
		return data, nil
	}

	median, err := data.Median()
	if err != nil {
		return nil, err
	}
	zthresh := 3.5
	adjust := 0.6745

	filtered := make(stats.Float64Data, 0, len(data))
	for _, d := range data {
		zscore := adjust * math.Abs(d-median) / mad
		if zscore < zthresh {
			filtered = append(filtered, d)
		}
	}

	return filtered, nil
}
