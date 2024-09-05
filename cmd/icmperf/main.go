package main

import (
	"time"

	"github.com/alecthomas/kong"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/prometheus-community/pro-bing"

	"github.com/b4nst/icmperf/pkg/cli"
	"github.com/b4nst/icmperf/pkg/model"
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
		cli.Duration = time.Duration(cli.Count * int(time.Second))
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
	m := model.NewModel(s, cli.Target, p.IPAddr().String(), cli.Duration+time.Second)

	go func() {
		s.Run()
	}()
	if _, err := tea.NewProgram(m).Run(); err != nil {
		ktx.FatalIfErrorf(err)
	}
}
