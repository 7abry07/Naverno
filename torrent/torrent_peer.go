package torrent

import (
	"Naverno/internal/bitfield"
	"Naverno/internal/choker"
	"Naverno/internal/peer"
	"Naverno/internal/peerprotocol"
	"maps"
	"net"
	"net/netip"
	"slices"
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

func (t *Torrent) handleChokerEvent(ev any) {
	switch ev.(type) {
	case choker.Unchoke:
		peers := slices.Collect(maps.Keys(t.peers))
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
		peers := slices.Collect(maps.Keys(t.peers))
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

func (t *Torrent) handlePeerMessage(pe peer.PeerMessage) {
	switch mess := pe.Message.(type) {
	case peerprotocol.Choke:
		{
			pe.AmChoked = true
			if d, ok := t.downloaders[pe.Peer]; ok {
				delete(t.downloaders, pe.Peer)
				t.stalledDownloaders[d.Piece] = d
				t.picker.SetFree(d.Piece)
				d.OnPeerChoke()
				t.logger.Info("torrent -> downloader stalled", "Piece", d.Piece.Idx)
			}
		}
	case peerprotocol.Unchoke:
		{
			pe.AmChoked = false
			t.download(pe.Peer)
		}
	case peerprotocol.Interested:
		{
			pe.AmInteresting = true
		}
	case peerprotocol.Uninterested:
		{
			pe.AmInteresting = false
		}
	case peerprotocol.Have:
		{
			if (pe.Pieces == bitfield.Bitfield{}) {
				pe.Pieces = bitfield.New(uint32(t.meta.PieceCount))
			}
			if mess.Idx > uint32(t.meta.PieceCount-1) {
				t.logger.Info("torrent -> invalid HAVE", "PeerID", string(pe.ID[:]), "Error", "Piece index out of bounds")
				t.closePeer(pe.Peer)
				return
			}

			pe.Pieces.Set(uint(mess.Idx))
			if pe.Pieces.Difference(t.bitset.BitSet).Any() && !pe.IsInteresting {
				pe.Interested()
			}
			t.picker.OnPeerHave(t.pieces[mess.Idx])
			t.download(pe.Peer)
		}
	case peerprotocol.Bitfield:
		{
			if (pe.Pieces != bitfield.Bitfield{}) {
				t.closePeer(pe.Peer)
				return
			}
			data, err := bitfield.From(mess.Pieces, uint32(t.meta.PieceCount))
			if err != nil {
				t.logger.Info("torrent -> invalid BITFIELD", "PeerID", string(pe.ID[:]), "Error", err)
				t.closePeer(pe.Peer)
				return
			}

			pe.Pieces = data
			if data.Difference(t.bitset.BitSet).Any() && !pe.IsInteresting {
				pe.Interested()
			}
			t.picker.OnPeerBitfield(pe)
		}
	case peerprotocol.Request:
	case peerprotocol.Piece:
		{
			downloader, ok := t.downloaders[pe.Peer]
			if !ok {
				return
			}

			pe.Uploaded += uint64(len(mess.Data))
			downloader.OnBlockReceived(mess.Begin, uint32(len(mess.Data)))
			t.writePiece(downloader.Piece, mess.Begin, mess.Data)
			if downloader.Completed() {
				t.pieceCompleted(downloader.Piece)
				delete(t.downloaders, pe.Peer)
			}

			t.download(pe.Peer)
		}
	case peerprotocol.Cancel:
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
