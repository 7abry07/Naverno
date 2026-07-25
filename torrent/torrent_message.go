package torrent

import (
	"Naverno/internal/bitfield"
	"Naverno/internal/peer"
	"Naverno/internal/peerprotocol"
	"Naverno/internal/requesthandler"
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
			replace := t.choker.OnInterested(pe)
			if replace != nil {
				replace.(*peer.Peer).Choke()
			}
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
		{
			handler := requesthandler.New(pe.ID, t.storage, t.pieces[mess.Idx], mess)
			t.requestHandlers[mess] = handler
			go handler.Run(t.requestHandlersResults)
			t.logger.Info("torrent -> started request handler", "Peer", string(pe.ID[:]), "Piece", mess.Idx, "Block", mess.Begin)
		}
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
		{
			delete(t.requestHandlers, peerprotocol.Request{Idx: mess.Idx, Begin: mess.Begin, Length: mess.Length})
			t.logger.Info("torrent -> request canceled", "Peer", string(pe.ID[:]), "Piece", mess.Idx, "Block", mess.Begin)
		}
	}
}
