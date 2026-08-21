package metadata

import (
	"Naverno/cmd/tui/utils"
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
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true).
		Padding(0, 1)

	if m.torrent == nil {
		return style.Render(m.viewport.View())
	}

	meta, _ := m.torrent.Metadata()
	b := &strings.Builder{}
	fmt.Fprintf(b, "Name           %v\n", meta.Name)
	fmt.Fprintf(b, "Length         %v\n", utils.FormatLength(uint64(meta.Length)))
	fmt.Fprintf(b, "Private        %v\n", meta.Private)
	fmt.Fprintf(b, "Pieces         %v\n", meta.PieceCount)
	fmt.Fprintf(b, "Piece Length   %v\n", utils.FormatLength(uint64(meta.PieceLength)))
	fmt.Fprintf(b, "Info Hash      %x\n", meta.Infohash)
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
