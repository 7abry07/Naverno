package torrent

import (
	"Naverno/internal/bitfield"
	"Naverno/internal/peer"
	"Naverno/internal/peerprotocol"
	"fmt"
)

func (t *Torrent) handlePeerMessage(pe peer.PeerMessage) {
	switch mess := pe.Message.(type) {
	case peerprotocol.Choke:
		{
			t.logger.Debug("torrent -> received CHOKE", "PeerID", string(pe.ID[:]))
			pe.AmChoked = true

			if d, ok := t.downloaders[pe.Peer]; ok {
				delete(t.downloaders, pe.Peer)
				t.stalledDownloaders[d.Piece] = d
				t.picker.OnPieceStalled(d.Piece)
				d.OnPeerChoke()
				t.logger.Info("torrent -> downloader stalled", "Piece", d.Piece)
			}
		}
	case peerprotocol.Unchoke:
		{
			t.logger.Debug("torrent -> received UNCHOKE", "PeerID", string(pe.ID[:]))
			pe.AmChoked = false
			if pe.Pieces != nil {
				t.download(pe.Peer)
			}
		}
	case peerprotocol.Interested:
		{
			t.logger.Debug("torrent -> received INTERESTED", "PeerID", string(pe.ID[:]))
			pe.AmInteresting = true
		}
	case peerprotocol.Uninterested:
		{
			t.logger.Debug("torrent -> received UNINTERESTED", "PeerID", string(pe.ID[:]))
			pe.AmInteresting = false
		}
	case peerprotocol.Have:
		{
			t.logger.Debug("torrent -> received HAVE", "PeerID", string(pe.ID[:]), "Idx", mess.Idx)
			if pe.Pieces == nil {
				pe.Pieces = bitfield.New(uint32(t.meta.PieceCount))
			}
			if mess.Idx > uint32(t.meta.PieceCount-1) {
				t.logger.Info("torrent -> invalid HAVE", "PeerID", string(pe.ID[:]), "Error", "Piece index out of bounds")
				t.closePeer(pe.Peer)
				return
			}

			pe.Pieces.Set(mess.Idx)
			t.picker.OnPeerHave(mess.Idx)
			if !t.pieces.Test(mess.Idx) {
				pe.IsInteresting = true
			}

			if !pe.AmChoked {
				t.download(pe.Peer)
			}

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

			for i := range data.SetBits() {
				if !t.pieces.Test(uint32(i)) {
					pe.IsInteresting = true
					break
				}
			}

			pe.Pieces = data
			t.logger.Debug("torrent -> received BITFIELD", "PeerID", string(pe.ID[:]), "Pieces", pe.Pieces.Count())
			t.picker.OnPeerBitfield(pe)
		}
	case peerprotocol.Request:
		t.logger.Debug("torrent -> received REQUEST", "PeerID", string(pe.ID[:]), "Request", fmt.Sprintf("%v, %v, %v", mess.Idx, mess.Begin, mess.Length))
	case peerprotocol.Piece:
		{
			t.logger.Debug("torrent -> received PIECE", "PeerID", string(pe.ID[:]), "Block", fmt.Sprintf("%v, %v, %v", mess.Idx, mess.Begin, len(mess.Data)))
			downloader, ok := t.downloaders[pe.Peer]
			if !ok {
				return
			}

			downloader.OnBlockReceived(mess.Begin, uint32(len(mess.Data)))
			if downloader.Completed() {
				t.pieceCompleted(downloader, pe.Peer)
			}

			if t.pieces.All() {
				t.announceCompleted()
				t.logger.Info("torrent -> completed")
				return
			}

			if !pe.AmChoked {
				t.download(pe.Peer)
			}
		}
	case peerprotocol.Cancel:
		t.logger.Info("torrent -> received CANCEL", "PeerID", string(pe.ID[:]), "Idx", mess.Idx, "Begin", mess.Begin, "Length", mess.Length)
	}
}
