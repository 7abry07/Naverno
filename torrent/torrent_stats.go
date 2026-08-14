package torrent

import (
	"Naverno/internal/tracker"
	"net"
	"time"
)

type PeerInfo struct {
	ID           [20]byte
	Address      net.Addr
	DownloadRate uint64
	UploadRate   uint64
	Downloaded   uint64
	Uploaded     uint64
}

type TrackerInfo struct {
	URL          string
	NextAnnounce time.Time
	MinInterval  time.Time
	Seeders      int64
	Leechers     int64
	Error        error
}

type PieceInfo struct {
	Completed    bool
	Availability uint64
}

type TorrentStats struct {
	PiecesDownloaded uint64
	Downloaded       uint64
	Uploaded         uint64
	Connections      int
	Peers            []PeerInfo
	Pieces           []PieceInfo
	Trackers         []TrackerInfo
}

func (t *Torrent) GetStats() *TorrentStats {
	t.statsRequest <- TorrentStats{}
	stats := <-t.statsRequest
	return &stats
}

func (t *Torrent) fillStats(req *TorrentStats) {
	req.Downloaded = uint64(t.downloaded)
	req.Uploaded = uint64(t.uploaded)
	req.PiecesDownloaded = uint64(t.bitset.Count())
	req.Connections += len(t.peers) + len(t.outgoing)
	for id, p := range t.peers {
		req.Peers = append(req.Peers, PeerInfo{
			ID:           id,
			Address:      p.Addr(),
			DownloadRate: p.DownloadRate(),
			UploadRate:   p.UploadRate(),
			Downloaded:   p.Downloaded(),
			Uploaded:     p.Uploaded(),
		})
	}

	for url, stats := range t.announcer.GetTrackerStats() {
		tr := TrackerInfo{
			URL: url,
		}
		switch stats := stats.(type) {
		case error:
			tr.Error = stats
		case *tracker.AnnounceResponse:
			tr.NextAnnounce = time.Now().Add(stats.Interval)
			tr.MinInterval = time.Now().Add(stats.MinInterval)
			tr.Leechers = stats.Leechers
			tr.Seeders = stats.Seeders
		}
		req.Trackers = append(req.Trackers, tr)
	}

	for idx, avl := range t.picker.GetAvailability() {
		req.Pieces = append(req.Pieces, PieceInfo{
			Availability: uint64(avl),
			Completed:    t.bitset.Test(uint(idx)),
		})
	}

}
