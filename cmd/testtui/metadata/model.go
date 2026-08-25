package metadata

import (
	"Naverno/cmd/testtui/utils"
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
	Style    lipgloss.Style
}

func New(w, h int) Model {
	return Model{
		viewport: viewport.New(viewport.WithWidth(w), viewport.WithHeight(h)),
	}
}

func (m Model) Init() tea.Cmd {
	return m.viewport.Init()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	b := &strings.Builder{}
	m.viewport.Style = m.Style
	b.WriteString("Metadata\n\n")

	if m.torrent != nil {
		meta := m.torrent.Metadata()
		fmt.Fprintf(b, "Name           %v\n", meta.Name)
		fmt.Fprintf(b, "Length         %v\n", utils.FormatLength(uint64(meta.Info.Length)))
		fmt.Fprintf(b, "Private        %v\n", meta.Info.Private)
		fmt.Fprintf(b, "Pieces         %v\n", meta.Info.PieceCount)
		fmt.Fprintf(b, "Piece Length   %v\n", utils.FormatLength(uint64(meta.Info.PieceLength)))
		fmt.Fprintf(b, "Info Hash      %x\n", meta.Infohash)
		fmt.Fprintf(b, "Created By     %v\n", meta.CreatedBy)
		fmt.Fprintf(b, "Creation Date  %v\n", meta.CreatedAt)
		fmt.Fprintf(b, "Comment        %v\n", meta.Comment)
	}

	m.viewport.SetContent(b.String())
	return m.viewport.View()
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
