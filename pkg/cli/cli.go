package cli

import (
	"fmt"
	"time"
)

type DataSizes []int

func (ds *DataSizes) Validate() error {
	if len(*ds) == 0 {
		return fmt.Errorf("data sizes must have at least one element")
	}

	for _, size := range *ds {
		if size < 24 {
			return fmt.Errorf("data size must be greater than 0")
		}
	}
	return nil
}

type CLI struct {
	Timeout    time.Duration `help:"The timeout for each ping." short:"t" default:"5s"`
	Duration   time.Duration `help:"The duration of the test." short:"d" xor:"duration"`
	Count      int           `help:"The number of pings to send." short:"c" default:"10" xor:"duration"`
	Privileged bool          `help:"Use privileged ICMP sockets." default:"false" short:"p"`
	DataSizes  []int         `help:"The size of the data to send." short:"s" default:"1024"`

	Target string `arg:"" help:"The target host to ping."`
}
