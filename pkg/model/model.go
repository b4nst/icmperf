package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/b4nst/icmperf/pkg/session"
)

var (
	maxWidth = 100
)

type Model struct {
	session  *session.Session
	duration time.Duration
	peer     string
	peerAddr string

	timer    timer.Model
	progress progress.Model
	stats    *session.Statistics
}

func NewModel(sess *session.Session, peer string, peerAddr string, duration time.Duration) *Model {
	return &Model{
		session:  sess,
		duration: duration,
		peer:     peer,
		peerAddr: peerAddr,

		timer:    timer.NewWithInterval(duration, 200*time.Millisecond),
		progress: progress.New(progress.WithFillCharacters('▱', ' '), progress.WithDefaultGradient()),
		stats:    new(session.Statistics),
	}
}

type statsMsg *session.Statistics
type statsErr error

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := []tea.Cmd{}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			cmds = append(cmds, tea.Quit)
		}
	case timer.TimeoutMsg:
		cmds = append(cmds, tea.Quit)
	case timer.TickMsg:
		var tc, pc tea.Cmd
		m.timer, tc = m.timer.Update(msg)
		pc = m.progress.SetPercent(1.0 - m.timer.Timeout.Seconds()/m.duration.Seconds())
		cmds = append(cmds, tc, pc)
	case tea.WindowSizeMsg:
		m.progress.Width = min(msg.Width-4, maxWidth)
	case statsMsg:
		*m.stats = *msg
		cmds = append(cmds, m.statsCmd())
	case statsErr:
		cmds = append(cmds, m.statsCmd())
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	view := strings.Builder{}

	view.WriteString(fmt.Sprintf("Connecting to host %s (%s)\n", m.peer, m.peerAddr))

	view.WriteString("ETC: ")
	view.WriteString(m.timer.View())
	view.WriteString("  ")
	view.WriteString(m.progress.View())
	view.WriteString("\n\n")

	if m.stats != nil {
		view.WriteString(m.stats.String())
	}

	return view.String()
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.timer.Init(), m.statsCmd())
}

func (m *Model) statsCmd() tea.Cmd {
	return func() tea.Msg {
		stats, err := m.session.Statistics()
		if err != nil {
			return statsErr(err)
		}
		return statsMsg(stats)
	}
}
