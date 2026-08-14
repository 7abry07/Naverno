package torrent

import (
	"Naverno/internal/bitfield"
	"Naverno/internal/peer"
	"Naverno/internal/peerprotocol"
	"Naverno/internal/piecedownloader"
	"fmt"
)

func (t *Torrent) handlePeerMessage(pe peer.PeerMessage) {
	switch mess := pe.Message.(type) {
	case peerprotocol.Choke:
		{
			pe.AmChoked = true
			if d, ok := t.downloaders[pe.Peer]; ok {
				delete(t.downloaders, pe.Peer)
				t.stalledDownloaders[d.Piece] = d
				t.picker.SetFree(d.Piece.Idx)
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
			if pe.Pieces == nil {
				pe.Pieces = bitfield.New(uint32(t.meta.PieceCount))
			}
			if mess.Idx > uint32(t.meta.PieceCount-1) {
				t.logger.Info("torrent -> invalid HAVE", "PeerID", string(pe.ID[:]), "Error", "Piece index out of bounds")
				t.closePeer(pe.Peer)
				return
			}

			pe.Pieces.Set(uint(mess.Idx))
			if pe.Pieces.Difference(t.bitset.BitSet).Any() && !pe.IsInteresting {
				pe.Interesting()
			}
			t.picker.OnPeerHave(mess.Idx)
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
			if data.Difference(t.bitset.BitSet).Any() && !pe.IsInteresting {
				pe.Interesting()
			}
			t.picker.OnPeerBitfield(data)
		}
	case peerprotocol.Request:
		{
			t.pendingRequests[mess] = pe.ID
			t.storage.AsyncRead(t.readResults, t.pieces[mess.Idx], mess.Begin, mess.Length)
			t.logger.Debug("torrent -> started request handler", "Peer", string(pe.ID[:]), "Piece", mess.Idx, "Block", mess.Begin)
		}
	case peerprotocol.Piece:
		{
			downloader, ok := t.downloaders[pe.Peer]
			if !ok {
				return
			}

			if mess.Idx > uint32(len(t.pieces)-1) {
				t.logger.Info("torrent -> invalid PIECE", "PeerID", string(pe.ID[:]), "Error", "Invalid piece index")
				t.closePeer(pe.Peer)
				return
			}

			err := downloader.OnBlockReceived(mess.Begin, uint32(len(mess.Data)))
			switch err {
			case piecedownloader.ErrInvalid:
				t.logger.Info("downloader -> invalid block", "PeerID", string(pe.ID[:]), "Block", fmt.Sprintf("(%v, %v, %v)", mess.Idx, mess.Begin, len(mess.Data)))
				t.closePeer(pe.Peer)
				return
			case piecedownloader.ErrDuplicate:
				t.logger.Info("downloader -> duplicate block", "PeerID", string(pe.ID[:]), "Block", fmt.Sprintf("(%v, %v, %v)", mess.Idx, mess.Begin, len(mess.Data)))
				return
			case piecedownloader.ErrNotRequested:
				t.logger.Info("downloader -> not requested block", "PeerID", string(pe.ID[:]), "Block", fmt.Sprintf("(%v, %v, %v)", mess.Idx, mess.Begin, len(mess.Data)))
			case nil:
			default:
				t.closePeer(pe.Peer)
				return
			}

			t.writePiece(downloader.Piece, mess.Begin, mess.Data)
			pe.UpdateStats(uint64(len(mess.Data)), 0)
			if downloader.Completed() {
				t.pieceCompleted(downloader.Piece)
				delete(t.downloaders, pe.Peer)
			}

			t.download(pe.Peer)
		}
	case peerprotocol.Cancel:
		{
			delete(t.pendingRequests, peerprotocol.Request{Idx: mess.Idx, Begin: mess.Begin, Length: mess.Length})
			t.logger.Info("torrent -> request canceled", "Peer", string(pe.ID[:]), "Piece", mess.Idx, "Block", mess.Begin)
		}
	}
}
