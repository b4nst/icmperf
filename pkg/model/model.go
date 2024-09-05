package model

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dustin/go-humanize"

	"github.com/b4nst/icmperf/pkg/recorder"
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
	table    table.Model
}

func NewModel(session *session.Session, peer string, peerAddr string, duration time.Duration) *Model {
	return &Model{
		session:  session,
		duration: duration,
		peer:     peer,
		peerAddr: peerAddr,

		timer:    timer.NewWithInterval(duration, 200*time.Millisecond),
		progress: progress.New(progress.WithFillCharacters('▱', ' '), progress.WithDefaultGradient()),
		table: table.New(table.WithFocused(false), table.WithColumns([]table.Column{
			{Title: "Seq", Width: 10},
			{Title: "Latency", Width: 10},
			{Title: "RTT", Width: 10},
			{Title: "Bitrate", Width: 20},
		})),
	}
}

type statsMsg []table.Row
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
		var cmd tea.Cmd
		m.table.SetRows(msg)
		m.table.SetHeight(len(msg) + 2)
		m.table.GotoBottom()
		m.table, cmd = m.table.Update(msg)
		cmds = append(cmds, cmd, m.statsCmd())
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

	view.WriteString(m.table.View())

	return view.String()
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.timer.Init(), m.statsCmd())
}

func (m *Model) statsCmd() tea.Cmd {
	return func() tea.Msg {
		stats := m.session.Statistics()
		avg, err := recorder.ProcessStats(stats)
		if err != nil {
			return statsErr(err)
		}
		// Sort by sequence number
		sort.Sort(recorder.BySeq(stats))
		rows := make([]table.Row, 0, len(stats))
		for _, stat := range stats {
			rows = append(rows, table.Row{
				fmt.Sprintf("%d", stat.Seq),
				fmt.Sprintf("%s", stat.Latency.Truncate(time.Millisecond)),
				fmt.Sprintf("%s", stat.Rtt.Truncate(time.Millisecond)),
				fmt.Sprintf("%s/sec", humanize.Bytes(uint64(stat.Bandwidth))),
			})
		}
		rows = append(rows, table.Row{
			"Average",
			fmt.Sprintf("%s", avg.Latency.Truncate(time.Millisecond)),
			fmt.Sprintf("%s", avg.Rtt.Truncate(time.Millisecond)),
			fmt.Sprintf("%s/sec", humanize.Bytes(uint64(avg.Bandwidth))),
		})

		return statsMsg(rows)
	}
}
