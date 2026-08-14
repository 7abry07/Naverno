package torrent

import (
	"Naverno/internal/choker"
	"Naverno/internal/peer"
	"maps"
	"net"
	"net/netip"
	"slices"
	"time"
)

func (t *Torrent) handleDisconnected(p *peer.Peer) {
	t.picker.OnPeerDisconnected(p.Pieces)
	downloader, ok := t.downloaders[p]
	if ok {
		t.stalledDownloaders[downloader.Piece] = downloader
		t.picker.SetFree(downloader.Piece.Idx)
		downloader.OnPeerDisconnected()
		t.logger.Info("torrent -> downloader stalled", "Piece", downloader.Piece.Idx)
	}
	t.closePeer(p)
	t.logger.Info("torrent -> peer disconnected", "Address", p.Addr().String(), "Peer", string(p.ID[:]))
}

func (t *Torrent) handleChokerEvent(ev any) {
	switch ev.(type) {
	case choker.Unchoke:
		peers := slices.Collect(maps.Values(t.peers))
		slice := make([]choker.Peer, len(peers))
		for i, p := range peers {
			slice[i] = choker.Peer(p)
		}
		pickedSlice := t.choker.PickPeers(slice)
		picked := make([]*peer.Peer, len(pickedSlice))
		for i, p := range pickedSlice {
			picked[i] = p.(*peer.Peer)
		}

		t.logger.Debug("choker -> unchoke event", "Peers", len(picked))

		for _, p := range peers {
			if slices.Contains(picked, p) {
				p.Unchoke()
			}
			p.Choke()
		}
	case choker.Optimistic:
		peers := slices.Collect(maps.Values(t.peers))
		slice := make([]choker.Peer, len(peers))
		for i, p := range peers {
			slice[i] = choker.Peer(p)
		}
		optimistic := t.choker.PickOptimistic(slice)
		if optimistic == nil {
			return
		}
		p := optimistic.(*peer.Peer)
		t.logger.Debug("choker -> optimistic unchoke event", "Peer", string(p.ID[:]))
		p.Unchoke()
	}
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
