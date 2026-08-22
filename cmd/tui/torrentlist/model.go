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
	11,
	13,
	17,
	15,
	17,
}

var TableColumnFields = []string{
	"Name",
	"Length",
	"Status",
	"Progress",
	"Download Rate",
	"Upload Rate",
}

type StatsMsg struct {
	stats map[*TorrentEntry]*torrent.TorrentStats
}

func (l *Model) stats() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		stats := map[*TorrentEntry]*torrent.TorrentStats{}
		for _, e := range l.list {
			s := e.Handle.GetStats()
			stats[e] = s
		}
		return StatsMsg{stats}
	})
}

type Model struct {
	viewport      viewport.Model
	list          []*TorrentEntry
	SelectedStyle lipgloss.Style
	yOffset       int
	selected      int
	limits        []int
}

func New(w, h int) Model {
	limits := []int{}

	for _, limit := range TableLengthLimits {
		limits = append(limits, int((float64(limit)/100.0)*float64(w)))
	}
	v := viewport.New(viewport.WithWidth(w), viewport.WithHeight(h))

	return Model{
		limits:        limits,
		selected:      -1,
		yOffset:       0,
		SelectedStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#898486")),
		viewport:      v,
	}
}

func (l Model) Init() tea.Cmd {
	return l.stats()
}

func (l Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case StatsMsg:
		cmds := []tea.Cmd{}
		for i, e := range l.list {
			m, _ := e.Handle.Metadata()
			s, ok := msg.stats[e]
			if !ok {
				continue
			}
			if s.Error != nil {
				cmds = append(cmds, e.Progress.SetPercent(0))
				e.Errored(s.Error)
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
				e.Completed()
			}
			i++
		}

		cmds = append(cmds, l.stats())
		return l, tea.Batch(cmds...)
	case progress.FrameMsg:
		for _, e := range l.list {
			var cmd tea.Cmd
			e.Progress, cmd = e.Progress.Update(msg)
			if cmd != nil {
				return l, cmd
			}
		}
		return l, nil
	case tea.KeyMsg:
		if len(l.list) == 0 {
			return l, nil
		}
		switch msg.String() {
		case "k", "up":
			switch l.selected {
			case -1:
				l.selected = 0
			case 0:
				l.yOffset = 0
			default:
				l.selected -= 1
				if l.selected < l.yOffset {
					l.yOffset -= l.yOffset - l.selected
				}
			}
		case "j", "down":
			if len(l.list) == 0 {
				return l, nil
			}
			switch l.selected {
			case -1:
				l.selected = 0
			case len(l.list) - 1:
			default:
				l.selected += 1
				if l.viewport.Height()-4-(l.selected+1)+l.yOffset <= 0 {
					l.yOffset = -(l.viewport.Height() - 4 - (l.selected + 1))
				}
			}
		}
		return l, nil
	default:
		return l, nil
	}
}

func (l Model) View() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false).
		BorderBottom(true)

	b := &strings.Builder{}
	for i, text := range TableColumnFields {
		fmt.Fprintf(b, "%-*s", l.limits[i]+2, text)
	}
	s := style.Render(b.String())
	b.Reset()
	b.WriteString(s + "\n")

	for i, e := range l.list {
		if i >= l.yOffset && i < len(l.list)+l.yOffset {
			if i == l.selected {
				fmt.Fprintf(b, "%v\n", e.Render(l.SelectedStyle, l.limits))
			} else {
				fmt.Fprintf(b, "%v\n", e.Render(lipgloss.NewStyle(), l.limits))
			}
		}
	}

	l.viewport.SetContent(b.String())
	l.viewport.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true).
		Padding(0, 1)

	return l.viewport.View()
}

func (l *Model) AddTorrent(t *torrent.Torrent) {
	m, _ := t.Metadata()
	e := NewTorrentEntry(t, m.Name, uint64(m.Length))
	e.Downloading()
	l.list = append(l.list, e)
}

func (l *Model) RemoveTorrent(t *torrent.Torrent) tea.Cmd {
	temp := []*TorrentEntry{}
	index := 0
	for i, e := range l.list {
		if e.Handle != t {
			temp = append(temp, e)
			continue
		}
		index = i
	}
	if l.selected != 0 && l.selected >= index {
		l.selected -= 1
	}

	l.list = temp
	return func() tea.Msg { t.Stop(); return nil }
}

func (l *Model) GetSelected() *torrent.Torrent {
	if l.selected >= 0 {
		if l.selected <= len(l.list)-1 {
			return l.list[l.selected].Handle
		}
	}
	return nil
}

func (l *Model) GetPeers(t *torrent.Torrent) []torrent.PeerInfo {
	if t == nil {
		return []torrent.PeerInfo{}
	}

	for _, e := range l.list {
		if e.Handle == t {
			return e.Stats.Peers
		}
	}
	return []torrent.PeerInfo{}
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
