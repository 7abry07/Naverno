package torrentlist

import (
	"Naverno/cmd/tui/utils"
	"Naverno/torrent"
	"fmt"

	"charm.land/bubbles/v2/progress"
	"charm.land/lipgloss/v2"
)

var (
	StatusDownloadingStyle = lipgloss.NewStyle().Foreground(lipgloss.BrightBlue)
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
		return "Seeding"
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
	Handle     *torrent.Torrent
	Name       string
	StatusType TorrentStatus
	Status     string
	Length     uint64
	Stats      torrent.TorrentStats
	Progress   progress.Model
}

func NewTorrentEntry(handle *torrent.Torrent, name string, length uint64) *TorrentEntry {
	return &TorrentEntry{
		Handle: handle,
		Name:   name,
		Length: length,
		Progress: progress.New(
			progress.WithoutPercentage(),
			progress.WithColors(lipgloss.White, lipgloss.White),
		),
	}
}

func (e *TorrentEntry) Errored(err error) {
	e.Status = ErroredStatus.String() + ": " + err.Error()
	e.StatusType = ErroredStatus
}

func (e *TorrentEntry) Completed() {
	e.Status = CompletedStatus.String()
	e.StatusType = CompletedStatus
}

func (e *TorrentEntry) Downloading() {
	e.Status = DownloadingStatus.String()
	e.StatusType = DownloadingStatus
}

func (e *TorrentEntry) Render(selected lipgloss.Style, limits []int) string {
	name := utils.Clamp(e.Name, limits[0])
	length := utils.Clamp(utils.FormatLength(e.Length), limits[1])
	status := utils.Clamp(e.Status, limits[2])
	e.Progress.SetWidth(limits[3])
	// peers := utils.Clamp(fmt.Sprintf("%v", len(e.Stats.Peers)), limits[4])
	drate := utils.Clamp(fmt.Sprintf("%v", utils.FormatRate(e.Stats.DownloadRate)), limits[4])
	urate := utils.Clamp(fmt.Sprintf("%v", utils.FormatRate(e.Stats.UploadRate)), limits[5])

	return fmt.Sprintf("%v  %v  %v  %v  %v  %v",
		selected.Render(name),
		length,
		e.StatusType.Style(status),
		e.Progress.View(),
		// peers,
		drate,
		urate,
	)
}
