package torrent

import "Naverno/internal/announcer"

func (t *Torrent) handleAnnounce() {
	t.torrentAnnounce <- announcer.Torrent{
		InfoHash:   t.infohash,
		PeerID:     t.session.pid,
		Downloaded: t.downloaded,
		Uploaded:   t.uploaded,
		Left:       t.left,
	}
}

func (t *Torrent) announceCompleted() {
	t.announcer.Completed(
		announcer.Torrent{
			InfoHash:   t.infohash,
			PeerID:     t.session.pid,
			Downloaded: t.downloaded,
			Uploaded:   t.uploaded,
			Left:       t.left,
		})
}

func (t *Torrent) AnnounceToAllTrackers() {
	if t.announcer != nil {
		t.announcer.AnnounceToAllTrackers(announcer.Torrent{
			InfoHash:   t.infohash,
			PeerID:     t.session.pid,
			Downloaded: t.downloaded,
			Uploaded:   t.uploaded,
			Left:       t.left,
		})
	}
}
