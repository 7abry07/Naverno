package torrent

import "Naverno/internal/announcer"

func (t *Torrent) handleAnnounce() {
	t.torrentAnnounce <- announcer.Torrent{
		InfoHash:   t.meta.Infohash,
		PeerID:     t.pid,
		Downloaded: t.downloaded,
		Uploaded:   t.uploaded,
		Left:       t.left,
	}
}

func (t *Torrent) closeAnnouncer() {
	t.announcer.Close()
	t.torrentAnnounce <- announcer.Torrent{
		InfoHash:   t.meta.Infohash,
		PeerID:     t.pid,
		Downloaded: t.downloaded,
		Uploaded:   t.uploaded,
		Left:       t.left,
	}
}
