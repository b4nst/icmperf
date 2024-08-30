package model

import (
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dustin/go-humanize"

	"github.com/b4nst/icmperf/pkg/pinger"
	"github.com/b4nst/icmperf/pkg/recorder"
)

var (
	maxWidth = 100
)

type Model struct {
	pinger  *pinger.Pinger
	record  *recorder.Record
	peer    *net.UDPAddr
	payload []byte
	stats   *recorder.Stats

	duration time.Duration
	timer    timer.Model
	progress progress.Model
}

func NewModel(pinger *pinger.Pinger, record *recorder.Record, peer *net.UDPAddr, mtu int, duration time.Duration) *Model {
	return &Model{
		pinger:  pinger,
		record:  record,
		peer:    peer,
		payload: make([]byte, mtu-28),
		stats:   nil,

		duration: duration,
		timer:    timer.NewWithInterval(duration, 200*time.Millisecond),
		progress: progress.New(progress.WithFillCharacters('▱', ' '), progress.WithDefaultGradient()),
	}
}

type latencyTick time.Time
type pingTick time.Time
type statsMsg *recorder.Stats

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := []tea.Cmd{}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			cmds = append(cmds, tea.Quit)
		}
	case timer.TickMsg:
		var tc, pc tea.Cmd
		m.timer, tc = m.timer.Update(msg)
		pc = m.progress.SetPercent(1.0 - m.timer.Timeout.Seconds()/m.duration.Seconds())
		cmds = append(cmds, tc, pc)
	case timer.TimeoutMsg:
		cmds = append(cmds, m.progress.SetPercent(1.0))
		cmds = append(cmds, m.statsCmd())
	case tea.WindowSizeMsg:
		m.progress.Width = min(msg.Width-4, maxWidth)
	case latencyTick:
		if !m.timer.Timedout() {
			cmds = append(cmds, m.pingLatency())
		}
	case pingTick:
		if !m.timer.Timedout() {
			cmds = append(cmds, m.pingBandwidth())
		}
	case statsMsg:
		m.stats = msg
		cmds = append(cmds, tea.Quit)

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	view := strings.Builder{}

	view.WriteString("Probing peer at ")
	view.WriteString(m.peer.IP.String())
	view.WriteString("\n\n")

	view.WriteString("Probe size: ")
	view.WriteString(humanize.Bytes(uint64(len(m.payload))))
	view.WriteString(" ")
	view.WriteString("ETC: ")
	view.WriteString(m.timer.View())
	view.WriteString("\n")

	view.WriteString(m.progress.View())
	view.WriteString("\n\n")

	if m.stats != nil {
		view.WriteString("Bandwidth: ")
		view.WriteString(humanize.Bytes(uint64(m.stats.Bandwidth())))
		view.WriteString("/s  ")

		view.WriteString("Latency: ")
		view.WriteString(m.stats.Latency().Round(time.Millisecond).String())
		view.WriteString("  ")

		view.WriteString("Loss rate: ")
		view.WriteString(strconv.FormatFloat(m.stats.PacketLoss()*100, 'f', 2, 64))
		view.WriteString("%\n")
	}

	return view.String()
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.pingLatency(), m.pingBandwidth(), m.timer.Init())
}

func (m *Model) pingLatency() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		sendAndRecord(m.pinger, m.peer, []byte{}, m.record)
		return latencyTick(t)
	})
}

func (m *Model) pingBandwidth() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		sendAndRecord(m.pinger, m.peer, m.payload, m.record)
		return pingTick(t)
	})
}

func (m *Model) statsCmd() tea.Cmd {
	return func() tea.Msg {
		return statsMsg(m.record.Stats())
	}
}

func sendAndRecord(p *pinger.Pinger, peer *net.UDPAddr, payload []byte, r *recorder.Record) error {
	id, t, err := p.Send(peer, payload)
	if err != nil {
		return err
	}
	r.PacketSent(id, len(payload), t)
	return nil
}
