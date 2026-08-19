package torrentlist

import (
	"Naverno/torrent"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var TableLengthLimits = []int{
	20,
	15,
	20,
	30,
	15,
}

var TableColumnFields = []string{
	"NAME",
	"LENGTH",
	"STATUS",
	"PROGRESS",
	"PEERS",
}

func (l *Model) stats() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		stats := map[*torrent.Torrent]*torrent.TorrentStats{}
		for t := range l.torrents {
			s := t.GetStats()
			stats[t] = s
		}
		return StatsMsg{stats}
	})
}

type Model struct {
	limits   []int
	viewport viewport.Model
	torrents map[*torrent.Torrent]*TorrentEntry
	list     []*TorrentEntry
}

func New(w, h int) Model {
	limits := []int{}

	for _, limit := range TableLengthLimits {
		limits = append(limits, int((float64(limit)/100.0)*float64(w)))
	}
	return Model{
		limits:   limits,
		torrents: make(map[*torrent.Torrent]*TorrentEntry),
		viewport: viewport.New(viewport.WithWidth(w), viewport.WithHeight(h)),
	}
}

func (l Model) Init() tea.Cmd {
	return l.stats()
}

func (l Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case AddTorrentMsg:
		t := msg.Torrent
		m, _ := t.Metadata()
		e := newTorrentEntry(m.Name, DownloadingStatus, uint64(m.Length))
		l.torrents[t] = e
		l.list = append(l.list, e)
		return l, nil
	case RemoveTorrentMsg:
		delete(l.torrents, msg.Torrent)
		return l, nil
	case StatsMsg:
		cmds := []tea.Cmd{}
		i := 0
		for t, e := range l.torrents {
			if i == len(msg.stats) {
				break
			}
			m, _ := t.Metadata()
			s, ok := msg.stats[t]
			if !ok {
				continue
			}
			if s.Error != nil {
				cmds = append(cmds, e.Progress.SetPercent(0))
				e.Error = s.Error
				e.Status = ErroredStatus
			} else {
				perc := float64(s.Downloaded) / float64(m.Length)
				if perc > 0.99 && m.Length > int64(s.Downloaded) {
					perc = 0.99
				}
				cmds = append(cmds, e.Progress.SetPercent(perc))
			}
			e.Length = uint64(m.Length)
			e.Stats = *s

			if m.Length == int64(s.Downloaded) {
				e.Status = CompletedStatus
			}
			i++
		}

		cmds = append(cmds, l.stats())
		return l, tea.Batch(cmds...)
	case progress.FrameMsg:
		for _, e := range l.torrents {
			var cmd tea.Cmd
			e.Progress, cmd = e.Progress.Update(msg)
			if cmd != nil {
				return l, cmd
			}
		}
		return l, nil
	default:
		return l, nil
	}
}

func (l Model) View() tea.View {
	b := &strings.Builder{}
	for i, text := range TableColumnFields {
		fmt.Fprintf(b, "%-*s", l.limits[i]+2, text)
	}
	b.WriteString("\n")

	for i, e := range l.list {
		if i == 0 {
			fmt.Fprintf(b, "%v\n", lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false).BorderTop(true).Render(e.Render(l.limits)))
		} else {
			fmt.Fprintf(b, "%v\n", e.Render(l.limits))
		}
	}

	l.viewport.SetContent(b.String())
	l.viewport.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true).
		Padding(0, 1)

	v := tea.NewView(l.viewport.View())
	return v
}

func (l *Model) SetWidth(width int) {
	limits := []int{}
	for _, limit := range TableLengthLimits {
		limits = append(limits, int((float64(limit)/100.0)*float64(width)))
	}
	l.limits = limits
	l.viewport.SetWidth(width)
}

func (l *Model) SetHeight(height int) {
	l.viewport.SetHeight(height)
}
