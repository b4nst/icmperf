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

	// Latenncy pinger
	platency, err := probing.NewPinger(cli.Target)
	ktx.FatalIfErrorf(err)
	if cli.Duration > 0 {
		platency.Timeout = cli.Duration
	} else {
		cli.Duration = time.Duration(cli.Count * int(time.Second))
		platency.Count = cli.Count
	}
	platency.SetPrivileged(cli.Privileged)
	platency.Size = 24 // Minimum size

	// Data pingers
	pdata, err := probing.NewPinger(cli.Target)
	ktx.FatalIfErrorf(err)
	if cli.Duration > 0 {
		pdata.Timeout = cli.Duration
	} else {
		pdata.Count = cli.Count
	}
	pdata.Size = cli.DataSize
	pdata.SetPrivileged(cli.Privileged)

	pingers := []*probing.Pinger{platency, pdata}

	s := session.NewSession(pingers)
	m := model.NewModel(s, cli.Target, platency.IPAddr().String(), cli.Duration+time.Second)

	go func() {
		s.Run()
	}()
	if _, err := tea.NewProgram(m).Run(); err != nil {
		ktx.FatalIfErrorf(err)
	}
}
