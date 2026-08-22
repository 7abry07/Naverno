package peerlist

import (
	"Naverno/cmd/tui/utils"
	"Naverno/torrent"
	"cmp"
	"slices"

	// "cmp"
	"fmt"
	// "slices"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var TableLengthLimits = []int{
	20,
	20,
	20,
	20,
	20,
}

var TableColumnFields = []string{
	"Address",
	"Download Rate",
	"Upload Rate",
	"Downloaded",
	"Uploaded",
}

type Model struct {
	viewport      viewport.Model
	peers         []torrent.PeerInfo
	Style         lipgloss.Style
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

	return Model{
		viewport:      viewport.New(viewport.WithWidth(w), viewport.WithHeight(h)),
		limits:        limits,
		selected:      -1,
		yOffset:       0,
		SelectedStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#898486")),
		peers:         []torrent.PeerInfo{},
	}
}

func (m Model) Init() tea.Cmd {
	return m.viewport.Init()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

func (m *Model) View() string {
	b := &strings.Builder{}
	m.viewport.Style = m.Style

	style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false).
		BorderBottom(true)

	for i, text := range TableColumnFields {
		fmt.Fprintf(b, "%-*s", m.limits[i]+2, text)
	}
	s := style.Render(b.String())
	b.Reset()
	b.WriteString(s + "\n")

	for i, e := range m.peers {
		if i >= m.yOffset && i < len(m.peers)+m.yOffset {
			selected := i == m.selected
			fmt.Fprintf(b, "%v\n", m.renderPeer(e, selected))
		}
	}

	m.viewport.SetContent(b.String())
	return m.viewport.View()
}

func (m *Model) renderPeer(p torrent.PeerInfo, selected bool) string {
	addr := utils.Clamp(p.Address.String(), m.limits[0])

	if selected {
		addr = m.SelectedStyle.Render(addr)
	}

	drate := utils.Clamp(utils.FormatRate(p.DownloadRate), m.limits[1])
	urate := utils.Clamp(utils.FormatRate(p.UploadRate), m.limits[2])
	d := utils.Clamp(utils.FormatLength(p.Downloaded), m.limits[3])
	u := utils.Clamp(utils.FormatLength(p.Uploaded), m.limits[4])

	return fmt.Sprintf("%v  %v  %v  %v  %v",
		addr, drate, urate, d, u,
	)
}

func (m *Model) SetPeers(peers []torrent.PeerInfo) {
	m.peers = peers
	slices.SortFunc(m.peers, func(a, b torrent.PeerInfo) int {
		return cmp.Compare(a.Address.String(), b.Address.String())
	})
}

func (m *Model) RemovePeer(peer torrent.PeerInfo) {
	temp := []torrent.PeerInfo{}
	for _, p := range m.peers {
		if p.ID == peer.ID {
			continue
		}
		temp = append(temp, p)
	}
	m.peers = temp
}

func (m *Model) SetWidth(w int) {
	limits := []int{}
	for _, limit := range TableLengthLimits {
		limits = append(limits, int((float64(limit)/100.0)*float64(w)))
	}
	m.limits = limits
	m.viewport.SetWidth(w)
}

func (m *Model) SetHeight(h int) {
	m.viewport.SetHeight(h)
}
