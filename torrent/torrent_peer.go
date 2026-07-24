package torrent

import (
	"Naverno/internal/bitfield"
	"Naverno/internal/peer"
	"Naverno/internal/peerprotocol"
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
			pe.IsInteresting = pe.Pieces.Difference(t.bitset.BitSet).Any()
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
			pe.IsInteresting = data.Difference(t.bitset.BitSet).Any()
			t.picker.OnPeerBitfield(pe)
		}
	case peerprotocol.Request:
	case peerprotocol.Piece:
		{
			downloader, ok := t.downloaders[pe.Peer]
			if !ok {
				return
			}

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
