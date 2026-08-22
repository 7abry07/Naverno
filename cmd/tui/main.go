package main

import (
	"Naverno/cmd/tui/metadata"
	"Naverno/cmd/tui/peerlist"
	"Naverno/cmd/tui/torrentlist"
	"Naverno/torrent"
	"fmt"
	"os"
	"strings"

	"log/slog"
	"net/http"
	_ "net/http/pprof"

	"charm.land/bubbles/v2/filepicker"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/term"
)

type model struct {
	session        *torrent.Session
	picker         filepicker.Model
	torrents       torrentlist.Model
	metadata       metadata.Model
	peers          peerlist.Model
	pickingFile    bool
	terminalWidth  int
	terminalHeight int
}

func newModel(s *torrent.Session) *model {
	W, H, err := term.GetSize(os.Stdin.Fd())
	if err != nil {
		fmt.Printf("could't get terminal size -> %v\n", err)
		return nil
	}
	p := filepicker.New()
	p.CurrentDirectory = "/home"
	p.AllowedTypes = []string{".torrent"}

	torrentlistW := (W / 3) * 2
	torrentlistH := int(float64(H) / 1.8)
	metadataW := W - torrentlistW
	peerlistH := H - torrentlistH

	return &model{
		session:  s,
		picker:   p,
		torrents: torrentlist.New(torrentlistW, torrentlistH),
		metadata: metadata.New(metadataW, torrentlistH),
		peers:    peerlist.New(W, peerlistH),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.torrents.Init(), m.picker.Init())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	cmds := []tea.Cmd{}

	m.torrents, cmd = m.torrents.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.pickingFile {
				m.pickingFile = false
				return m, nil
			}
			return m, tea.Quit
		case "n":
			m.pickingFile = true
		case "r":
			t := m.torrents.GetSelected()
			if t != nil {
				cmds = append(cmds, m.torrents.RemoveTorrent(t))
			}
		}
		if !m.pickingFile {
			return m, tea.Batch(cmds...)
		}
	case tea.WindowSizeMsg:
		m.terminalWidth = msg.Width
		m.terminalHeight = msg.Height

		torrentlistW := (msg.Width / 3) * 2
		torrentlistH := int(float64(msg.Height) / 1.8)
		metadataW := msg.Width - torrentlistW
		peerlistH := msg.Height - torrentlistH

		m.torrents.SetWidth(torrentlistW)
		m.metadata.SetWidth(metadataW)
		m.torrents.SetHeight(torrentlistH)
		m.metadata.SetHeight(torrentlistH)
		m.peers.SetWidth(msg.Width)
		m.peers.SetHeight(peerlistH)
	}

	m.picker, cmd = m.picker.Update(msg)
	cmds = append(cmds, cmd)

	if ok, path := m.picker.DidSelectFile(msg); ok {
		if t, err := m.session.AddTorrentFromFile(path, "/home/fabry/Downloads"); err == nil {
			m.torrents.AddTorrent(t)
		}
		m.pickingFile = false
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() tea.View {
	var v tea.View
	v.AltScreen = true

	if m.terminalWidth < 64 || m.terminalHeight < 24 {
		v.Content = TerminalTooSmall(64, 24, m.terminalWidth, m.terminalHeight)
		return v
	}

	if m.pickingFile {
		v.Content = m.picker.View()
		return v
	}

	selected := m.torrents.GetSelected()
	if selected != nil {
		m.metadata.SetTorrent(selected)
		m.peers.SetPeers(m.torrents.GetPeers(selected))
	}

	v.Content = lipgloss.JoinVertical(
		lipgloss.Top,
		lipgloss.JoinHorizontal(
			lipgloss.Bottom,
			m.metadata.View(),
			m.torrents.View()),
		m.peers.View())
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
	s := torrent.StartSession(slog.New(slog.DiscardHandler))
	m := newModel(s)
	if m == nil {
		return
	}
	p := tea.NewProgram(m)

	go http.ListenAndServe(":6060", nil)

	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
