package torrent

import (
	"Naverno/internal/bitfield"
	"Naverno/internal/peer"
	"Naverno/internal/peerprotocol"
)

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
			if pe.Pieces == nil {
				pe.Pieces = bitfield.New(uint32(t.meta.PieceCount))
			}
			if mess.Idx > uint32(t.meta.PieceCount-1) {
				t.logger.Info("torrent -> invalid HAVE", "PeerID", string(pe.ID[:]), "Error", "Piece index out of bounds")
				t.closePeer(pe.Peer)
				return
			}

			pe.Pieces.Set(mess.Idx)
			pe.IsInteresting = pe.Pieces.Difference(t.bitset).Any()
			t.picker.OnPeerHave(t.pieces[mess.Idx])
			t.download(pe.Peer)
		}
	case peerprotocol.Bitfield:
		{
			if pe.Pieces != nil {
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
			pe.IsInteresting = data.Difference(t.bitset).Any()
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
			if downloader.Completed() {
				t.pieceCompleted(downloader.Piece)
				delete(t.downloaders, pe.Peer)
			}

			if t.bitset.All() {
				t.closeSeeds()
				t.announceCompleted()
				t.logger.Info("torrent -> completed")
				return
			}

			t.download(pe.Peer)
		}
	case peerprotocol.Cancel:
	}
}
