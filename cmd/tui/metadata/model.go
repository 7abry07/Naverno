package metadata

import (
	"Naverno/torrent"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Model struct {
	viewport viewport.Model
	torrent  *torrent.Torrent
}

func New(w, h int) *Model {
	return &Model{
		viewport: viewport.New(viewport.WithWidth(w), viewport.WithHeight(h)),
	}
}

func (m Model) Init() tea.Cmd {
	return m.viewport.Init()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	b := &strings.Builder{}
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true).
		Padding(0, 1)

	if m.torrent == nil {
		fmt.Fprintf(b, "Name           \n")
		fmt.Fprintf(b, "Length         \n")
		fmt.Fprintf(b, "Pieces         \n")
		fmt.Fprintf(b, "Piece Length   \n")
		fmt.Fprintf(b, "Private        \n")
		fmt.Fprintf(b, "Created By     \n")
		fmt.Fprintf(b, "Creation Date  \n")
		fmt.Fprintf(b, "Comment        \n")
		m.viewport.SetContent(b.String())
		return style.Render(m.viewport.View())
	}
	meta, _ := m.torrent.Metadata()
	fmt.Fprintf(b, "Name           %v\n", meta.Name)
	fmt.Fprintf(b, "Length         %v\n", meta.Length)
	fmt.Fprintf(b, "Pieces         %v\n", meta.PieceCount)
	fmt.Fprintf(b, "Piece Length   %v\n", meta.PieceLength)
	fmt.Fprintf(b, "Private        %v\n", meta.Private)
	fmt.Fprintf(b, "Created By     %v\n", meta.CreatedBy)
	fmt.Fprintf(b, "Creation Date  %v\n", meta.CreationDate)
	fmt.Fprintf(b, "Comment        %v\n", meta.Comment)
	m.viewport.SetContent(b.String())

	return style.Render(m.viewport.View())
}

func (m *Model) SetWidth(w int) {
	m.viewport.SetWidth(w)
}

func (m *Model) SetHeight(h int) {
	m.viewport.SetHeight(h)
}

func (m *Model) SetTorrent(t *torrent.Torrent) {
	m.torrent = t
}
