package torrent

import (
	"Naverno/internal/peer"
	"net"
	"net/netip"
	"time"
)

func (t *Torrent) handleDisconnected(p *peer.Peer) {
	t.picker.OnPeerDisconnected(p)
	downloader, ok := t.downloaders[p]
	if ok {
		t.stalledDownloaders[downloader.Piece] = downloader
		t.picker.SetFree(downloader.Piece)
		downloader.OnPeerDisconnected()
		t.logger.Info("torrent -> downloader stalled", "Piece", downloader.Piece.Idx)
	}
	t.closePeer(p)
	t.logger.Info("torrent -> peer disconnected", "Address", p.Addr().String(), "Peer", string(p.ID[:]))
}

func (t *Torrent) dial(peers []netip.AddrPort) {
	for _, a := range peers {
		go func() {
			conn, err := net.DialTimeout("tcp", a.String(), time.Second*5)
			if err != nil {
				t.logger.Debug("torrent -> error in connecting to remote peer", "Address", a.Addr().String(), "Error", err.Error())
				return
			}
			t.newConns <- conn
		}()
	}
}

func (t *Torrent) closePeer(p *peer.Peer) {
	delete(t.downloaders, p)
	delete(t.peers, p)
	p.Stop()
}

func (t *Torrent) closeSeeds() {
	for p := range t.peers {
		if p.Pieces.All() {
			t.closePeer(p)
		}
	}
}

func (t *Torrent) closePeers() {
	for p := range t.peers {
		t.closePeer(p)
	}
}
