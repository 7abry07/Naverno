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

func formatRate(rate uint64) string {
	if rate > 1000000000000 {
		return fmt.Sprintf("%.2f Tbps", float64(rate)/1000000000000.0)
	}
	if rate > 1000000000 {
		return fmt.Sprintf("%.2f Gbps", float64(rate)/1000000000.0)
	}
	if rate > 1000000 {
		return fmt.Sprintf("%.2f Mbps", float64(rate)/1000000.0)
	}
	if rate > 1000 {
		return fmt.Sprintf("%.2f Kbps", float64(rate)/1000.0)
	}
	return fmt.Sprintf("%v Bps", rate)
}

func formatLength(length uint64) string {
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
	length := clamp(formatLength(e.Length), limits[1])
	status := clamp(e.Status.String(), limits[2])
	e.Progress.SetWidth(limits[3])
	e.Progress.ShowPercentage = false
	peers := clamp(fmt.Sprintf("%v", len(e.Stats.Peers)), limits[4])
	drate := clamp(fmt.Sprintf("%v", formatRate(e.Stats.DownloadRate)), limits[5])
	urate := clamp(fmt.Sprintf("%v", formatRate(e.Stats.UploadRate)), limits[6])

	return fmt.Sprintf("%v  %v  %v  %v  %v  %v  %v",
		name,
		length,
		e.Status.Style(status),
		e.Progress.View(),
		peers,
		drate,
		urate,
	)
}
