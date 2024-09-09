package main

import (
	"fmt"
	"net"
	"time"

	"github.com/alecthomas/kong"

	"github.com/b4nst/icmperf/pkg/cli"
	"github.com/b4nst/icmperf/pkg/pinger"
)

type result struct {
	stats *pinger.Statistics
	err   error
}

func main() {
	cli := cli.CLI{}
	ktx := kong.Parse(&cli)

	pinger := pinger.NewPinger(cli.Size, cli.Duration)
	fmt.Printf("Evaluating %s (etc %s)...\n", cli.Target.IP, cli.Duration)

	r := make(chan result)
	go func() {
		stats, err := pinger.Eval((*net.IPAddr)(&cli.Target))
		r <- result{stats, err}
	}()
	start := time.Now()
	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case t := <-ticker.C:
			etc := cli.Duration - t.Sub(start)
			if etc < 0 {
				etc = 0
			}
			stats := pinger.Stats()
			fmt.Printf("%s (etc %s)...\n", stats, etc.Round(time.Second))
		case res := <-r:
			ktx.FatalIfErrorf(res.err)
			fmt.Println(res.stats)
			return
		}
	}
}
