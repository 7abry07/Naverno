package torrent

import (
	"Naverno/internal/announcer"
	"Naverno/internal/peer"
)

func (t *Torrent) Stop() {
	close(t.closeC)
	<-t.doneC
	t.logger.Info("torrent -> stopped")
}

func (t *Torrent) closePeer(p *peer.Peer) {
	delete(t.downloaders, p)
	delete(t.peers, p.ID)
	p.Stop()
}

func (t *Torrent) closeSeeds() {
	for _, p := range t.peers {
		if p.Pieces.All() {
			t.closePeer(p)
		}
	}
}

func (t *Torrent) closePeers() {
	for _, p := range t.peers {
		t.closePeer(p)
	}
}

func (t *Torrent) closeHandshakes() {
	for hs := range t.outgoing {
		hs.Close()
	}
}

func (t *Torrent) closeAnnouncer() {
	t.announcer.Close(
		announcer.Torrent{
			InfoHash:   t.meta.Infohash,
			PeerID:     t.session.pid,
			Downloaded: t.downloaded,
			Uploaded:   t.uploaded,
			Left:       t.left,
		})
}
