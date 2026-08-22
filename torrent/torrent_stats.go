package torrent

import (
	"Naverno/internal/metadata"
	"Naverno/internal/tracker"
	"context"
	"net"
	"time"
)

type Metadata struct{ metadata.Metadata }

func (t *Torrent) Metadata() (Metadata, bool) {
	if t.meta != nil {
		return Metadata{*t.meta}, true
	}
	return Metadata{}, false
}

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
	DownloadRate     uint64
	UploadRate       uint64
	Connections      int
	Error            error
}

func (t *Torrent) Stats(ctx context.Context) *TorrentStats {
	for {
		select {
		case t.statsReq <- TorrentStats{}:
		case stats := <-t.statsReq:
			return &stats
		case <-t.closeC:
			return nil
		case <-ctx.Done():
			return nil
		}
	}
}

func (t *Torrent) Peers(ctx context.Context) []PeerInfo {
	for {
		select {
		case t.peersReq <- []PeerInfo{}:
		case peers := <-t.peersReq:
			return peers
		case <-t.closeC:
			return nil
		case <-ctx.Done():
			return nil
		}
	}
}

func (t *Torrent) Trackers(ctx context.Context) []TrackerInfo {
	for {
		select {
		case t.trackersReq <- []TrackerInfo{}:
		case trackers := <-t.trackersReq:
			return trackers
		case <-t.closeC:
			return nil
		case <-ctx.Done():
			return nil
		}
	}
}
func (t *Torrent) Pieces(ctx context.Context) []PieceInfo {
	for {
		select {
		case t.piecesReq <- []PieceInfo{}:
		case pieces := <-t.piecesReq:
			return pieces
		case <-t.closeC:
			return nil
		case <-ctx.Done():
			return nil
		}
	}
}

func (t *Torrent) fillRequest(req any) {
	switch r := req.(type) {
	case *TorrentStats:
		r.Downloaded = uint64(t.downloaded)
		r.Uploaded = uint64(t.uploaded)
		r.PiecesDownloaded = uint64(t.bitset.Count())
		r.Connections += len(t.peers) + len(t.outgoing)
		r.Error = t.err
		t.rateMut.Lock()
		r.DownloadRate = t.downloadRate
		r.UploadRate = t.uploadRate
		t.rateMut.Unlock()
		req = r
	case *[]PeerInfo:
		for id, p := range t.peers {
			*r = append(*r, PeerInfo{
				ID:           id,
				Address:      p.Addr(),
				DownloadRate: p.DownloadRate(),
				UploadRate:   p.UploadRate(),
				Downloaded:   p.Downloaded(),
				Uploaded:     p.Uploaded(),
			})
		}
		req = r
	case *[]TrackerInfo:
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
			*r = append(*r, tr)
		}
		req = r
	case *[]PieceInfo:
		for idx, avl := range t.picker.GetAvailability() {
			*r = append(*r, PieceInfo{
				Availability: uint64(avl),
				Completed:    t.bitset.Test(uint(idx)),
			})
		}
		req = r
	}
}
