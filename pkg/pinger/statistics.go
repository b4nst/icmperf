package pinger

import (
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
)

type Statistics struct {
	Sent       int
	Received   int
	Duration   time.Duration
	PacketSize int
}

// Bitrate returns the bitrate of the test in bytes per second.
func (s *Statistics) Bitrate() float64 {
	return float64(s.ByteReceived()) / s.Duration.Seconds()
}

func (s *Statistics) ByteSent() int {
	return s.Sent * s.PacketSize
}

func (s *Statistics) ByteReceived() int {
	return s.Received * s.PacketSize
}

// Loss returns the loss rate of the test as a percentage.
func (s *Statistics) Loss() float64 {
	return 1.0 - float64(s.Received)/float64(s.Sent)
}

func (s *Statistics) String() string {
	return fmt.Sprintf("%s sent, %s received in %s at %s/sec\n", humanize.Bytes(uint64(s.ByteSent())), humanize.Bytes(uint64(s.ByteReceived())), s.Duration, humanize.Bytes(uint64(s.Bitrate())))
}
