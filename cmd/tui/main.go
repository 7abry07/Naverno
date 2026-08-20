package main

import (
	"Naverno/cmd/tui/torrentlist"
	"Naverno/torrent"
	"fmt"
	"os"
	"strings"

	"log/slog"
	"net/http"
	_ "net/http/pprof"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/term"
)

type model struct {
	Torrents       torrentlist.Model
	terminalWidth  int
	terminalHeight int
}

func newModel() *model {
	W, H, err := term.GetSize(os.Stdin.Fd())
	if err != nil {
		fmt.Printf("could't get terminal size -> %v\n", err)
		return nil
	}
	return &model{
		Torrents: torrentlist.New(W, int(float64(H)/1.8)),
	}
}

func (m model) Init() tea.Cmd {
	return m.Torrents.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	cmds := []tea.Cmd{}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.terminalWidth = msg.Width
		m.terminalHeight = msg.Height
		m.Torrents.SetWidth(msg.Width)
		m.Torrents.SetHeight(int(float64(msg.Height) / 1.8))
	}

	m.Torrents, cmd = m.Torrents.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() tea.View {
	if m.terminalWidth < 64 || m.terminalHeight < 24 {
		v := tea.NewView(TerminalTooSmall(64, 24, m.terminalWidth, m.terminalHeight))
		v.AltScreen = true
		return v
	}
	v := tea.NewView(m.Torrents.View().Content)
	v.AltScreen = true
	return v
}

func TerminalTooSmall(minimumW, minimumH, w, h int) string {
	b := &strings.Builder{}
	style := lipgloss.NewStyle().
		Width(w).
		Height(h).
		Align(lipgloss.Center, lipgloss.Center)

	currWidth := fmt.Sprintf("%v", w)
	currHeight := fmt.Sprintf("%v", h)
	minWidth := fmt.Sprintf("%v", minimumW)
	minHeight := fmt.Sprintf("%v", minimumH)

	if w < minimumW {
		currWidth = lipgloss.NewStyle().Foreground(lipgloss.BrightRed).Render(currWidth)
	} else {
		currWidth = lipgloss.NewStyle().Foreground(lipgloss.BrightGreen).Render(currWidth)
	}
	if h < minimumH {
		currHeight = lipgloss.NewStyle().Foreground(lipgloss.BrightRed).Render(currHeight)
	} else {
		currHeight = lipgloss.NewStyle().Foreground(lipgloss.BrightGreen).Render(currHeight)
	}
	minWidth = lipgloss.NewStyle().Foreground(lipgloss.BrightGreen).Render(minWidth)
	minHeight = lipgloss.NewStyle().Foreground(lipgloss.BrightGreen).Render(minHeight)

	b.WriteString("Current terminal size\n")
	fmt.Fprintf(b, "width: %v  height: %v\n\n", currWidth, currHeight)

	b.WriteString("Minimum terminal size\n")
	fmt.Fprintf(b, "width: %v  height: %v", minWidth, minHeight)
	return style.Render(b.String())
}

func main() {
	m := newModel()
	if m == nil {
		return
	}
	p := tea.NewProgram(m)
	s := torrent.StartSession(slog.New(slog.DiscardHandler))

	t, err := s.AddTorrentFromFile("/home/fabry/Downloads/debian.torrent", "/home/fabry/Downloads")
	if err != nil {
		panic(err)
	}
	t1, err := s.AddTorrentFromFile("/home/fabry/Downloads/fedora.torrent", "/home/fabry/Downloads")
	if err != nil {
		panic(err)
	}

	go http.ListenAndServe(":6060", nil)
	go p.Send(torrentlist.AddTorrentMsg{Torrent: t})
	go p.Send(torrentlist.AddTorrentMsg{Torrent: t1})

	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
