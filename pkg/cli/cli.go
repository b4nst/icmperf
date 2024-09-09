package cli

import (
	"fmt"
	"net"
	"time"
)

type Target net.IPAddr

func (t *Target) UnmarshalText(text []byte) error {
	ip, err := net.LookupIP(string(text))
	if err != nil {
		return fmt.Errorf("lookup IP: %v", err)
	}

	t.IP = ip[0]
	return nil
}

type CLI struct {
	Duration time.Duration `help:"Evaluation duration." short:"t" default:"5s"`
	Size     int           `help:"The size of the data probe to send. If 0, computed using interface MTU." short:"s" default:"0"`

	Target Target `arg:"" help:"The target host to ping."`
}
