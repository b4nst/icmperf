package cli

import (
	"fmt"
	"net"
	"time"
)

type Target net.UDPAddr

func (t *Target) UnmarshalText(text []byte) error {
	ip, err := net.LookupIP(string(text))
	if err != nil {
		return fmt.Errorf("failed to lookup target: %w", err)
	}
	t.IP = ip[0]
	return nil
}

type CLI struct {
	Target   Target        `arg:"" help:"The target host to ping."`
	MTU      int           `help:"The maximum transmission unit of your interface." short:"m" default:"1500"`
	Timeout  time.Duration `help:"The timeout for each ping." short:"t" default:"5s"`
	Duration time.Duration `help:"The duration of the test." short:"d" default:"30s"`
	BindAddr string        `help:"The address to bind the ICMP listener to." default:"0.0.0.0" short:"l"`
}
