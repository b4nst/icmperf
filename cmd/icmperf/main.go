package main

import (
	"fmt"
	"net"

	"github.com/alecthomas/kong"

	"github.com/b4nst/icmperf/pkg/cli"
	"github.com/b4nst/icmperf/pkg/pinger"
)

func main() {
	cli := cli.CLI{}
	ktx := kong.Parse(&cli)

	pinger := pinger.NewPinger(cli.Size, cli.Duration)
	stats, err := pinger.Eval((*net.IPAddr)(&cli.Target))
	ktx.FatalIfErrorf(err)

	fmt.Println(stats)
}
