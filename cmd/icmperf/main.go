package main

import (
	"context"
	"net"
	"time"

	"github.com/alecthomas/kong"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/net/icmp"

	"github.com/b4nst/icmperf/pkg/cli"
	"github.com/b4nst/icmperf/pkg/model"
	"github.com/b4nst/icmperf/pkg/pinger"
	"github.com/b4nst/icmperf/pkg/recorder"
)

func main() {
	cli := cli.CLI{}
	ktx := kong.Parse(&cli)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pinger := pinger.NewPinger(cli.BindAddr, cli.MTU, cli.Timeout)
	record := recorder.NewRecord()
	pinger.OnRecv(func(m *icmp.Message, t time.Time) error {
		body := m.Body.(*icmp.Echo)
		id := uint64(body.ID)<<32 | uint64(body.Seq)
		record.PacketReceived(id, len(body.Data), t)
		return nil
	})

	peer := net.UDPAddr(cli.Target)
	m := model.NewModel(pinger, record, &peer, cli.MTU, cli.Duration)
	if err := pinger.Start(ctx); err != nil {
		ktx.FatalIfErrorf(err)
	}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		ktx.FatalIfErrorf(err)
	}
}
