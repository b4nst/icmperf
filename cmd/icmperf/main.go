package main

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/prometheus-community/pro-bing"

	"github.com/b4nst/icmperf/pkg/cli"
	"github.com/b4nst/icmperf/pkg/recorder"
	"github.com/b4nst/icmperf/pkg/session"
)

func main() {
	cli := cli.CLI{}
	ktx := kong.Parse(&cli)

	pingers := make([]*probing.Pinger, 0, len(cli.DataSizes)+1)

	// Latenncy pinger
	p, err := probing.NewPinger(cli.Target)
	ktx.FatalIfErrorf(err)
	if cli.Duration > 0 {
		p.Timeout = cli.Duration
	} else {
		p.Count = cli.Count
	}
	p.SetPrivileged(cli.Privileged)
	p.Size = 24 // Minimum size
	pingers = append(pingers, p)

	// Data pingers
	for _, ds := range cli.DataSizes {
		p, err := probing.NewPinger(cli.Target)
		ktx.FatalIfErrorf(err)
		if cli.Duration > 0 {
			p.Timeout = cli.Duration
		} else {
			p.Count = cli.Count
		}
		p.Size = ds
		p.SetPrivileged(cli.Privileged)
		pingers = append(pingers, p)
	}

	s := session.NewSession(pingers)
	ktx.FatalIfErrorf(s.Run())

	stats := s.Statistics()
	stat, err := recorder.ProcessStats(stats)
	ktx.FatalIfErrorf(err)

	for _, s := range stats {
		fmt.Printf("%s\n", s)
	}
	fmt.Println("- - - - - - - - - - - - - - - - - - - - - - - - -")
	fmt.Printf("%s\n", stat)
}
