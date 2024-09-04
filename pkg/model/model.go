package model

import (
	"fmt"
	"net"
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
	record  *recorder.Record
	payload []byte

	duration time.Duration
	timer    timer.Model
	progress progress.Model
}

func NewModel(pinger *pinger.Pinger, record *recorder.Record, peer *net.UDPAddr, mtu int, duration time.Duration) *Model {
	return &Model{
		record:  record,
		payload: make([]byte, mtu-28),

		duration: duration,
		timer:    timer.NewWithInterval(duration, 200*time.Millisecond),
		progress: progress.New(progress.WithFillCharacters('▱', ' '), progress.WithDefaultGradient()),
	}
}

type pingTick time.Time

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
	case tea.WindowSizeMsg:
		m.progress.Width = min(msg.Width-4, maxWidth)
	case error:
		fmt.Println(msg)
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

	view.WriteString("Probe size: ")
	view.WriteString(humanize.Bytes(uint64(len(m.payload))))
	view.WriteString(" ")
	view.WriteString("ETC: ")
	view.WriteString(m.timer.View())
	view.WriteString("\n")

	view.WriteString(m.progress.View())
	view.WriteString("\n\n")

	return view.String()
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.timer.Init())
}
