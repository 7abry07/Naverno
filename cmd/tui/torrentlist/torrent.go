package torrentlist

import (
	"Naverno/torrent"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/progress"
	"charm.land/lipgloss/v2"
)

type TableColumn string

var (
	StatusDownloadingStyle = lipgloss.NewStyle().Foreground(lipgloss.Cyan)
	StatusErroredStyle     = lipgloss.NewStyle().Foreground(lipgloss.BrightRed)
	StatusCompletedStyle   = lipgloss.NewStyle().Foreground(lipgloss.BrightGreen)
)

type TorrentStatus uint8

const (
	DownloadingStatus TorrentStatus = iota
	ErroredStatus
	CompletedStatus
)

func (s TorrentStatus) String() string {
	switch s {
	case DownloadingStatus:
		return "Downloading"
	case ErroredStatus:
		return "Errored"
	case CompletedStatus:
		return "Completed (Seeding)"
	}
	return ""
}

func (s TorrentStatus) Style(status string) string {
	switch s {
	case DownloadingStatus:
		return StatusDownloadingStyle.Render(status)
	case ErroredStatus:
		return StatusErroredStyle.Render(status)
	case CompletedStatus:
		return StatusCompletedStyle.Render(status)
	}
	return ""
}

type TorrentEntry struct {
	Name     string
	Status   TorrentStatus
	Length   uint64
	Stats    torrent.TorrentStats
	Progress progress.Model
	Error    error
}

func newTorrentEntry(name string, status TorrentStatus, length uint64) *TorrentEntry {
	return &TorrentEntry{
		Name:     name,
		Status:   status,
		Length:   length,
		Progress: progress.New(progress.WithColors(lipgloss.White, lipgloss.White)),
		Error:    nil,
	}
}

func clamp(text string, limit int) string {
	if len(text) == limit {
		return text
	} else if len(text) > limit-3 {
		text = text[:limit-3] + "..."
	} else {
		bd := &strings.Builder{}
		bd.WriteString(text)
		for range limit - len(text) {
			bd.WriteString(" ")
		}
		text = bd.String()
	}
	return text
}

func formatLength(length int) string {
	if length > 1000000000000 {
		return fmt.Sprintf("%.2f TiB", float64(length)/1000000000000.0)
	}
	if length > 1000000000 {
		return fmt.Sprintf("%.2f GiB", float64(length)/1000000000.0)
	}
	if length > 1000000 {
		return fmt.Sprintf("%.2f MiB", float64(length)/1000000.0)
	}
	if length > 1000 {
		return fmt.Sprintf("%.2f KiB", float64(length)/1000.0)
	}
	return fmt.Sprintf("%v B", length)
}

func (e *TorrentEntry) Render(limits []int) string {
	name := clamp(e.Name, limits[0])
	length := clamp(formatLength(int(e.Length)), limits[1])
	status := clamp(e.Status.String(), limits[2])
	e.Progress.SetWidth(limits[3])
	peers := clamp(fmt.Sprintf("%v", len(e.Stats.Peers)), limits[4])
	e.Progress.ShowPercentage = false

	return fmt.Sprintf("%v  %v  %v  %v  %v",
		name,
		length,
		e.Status.Style(status),
		e.Progress.View(),
		peers,
	)
}
