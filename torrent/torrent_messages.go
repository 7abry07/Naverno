package torrent

import (
	"Naverno/internal/peer"
	"Naverno/internal/peerprotocol"
	"Naverno/internal/piecedownloader"
	"Naverno/internal/util"

	"github.com/bits-and-blooms/bitset"
)

func (t *Torrent) handlePeerMessage(pe peer.PeerMessage) {
	switch mess := pe.Message.(type) {
	case peerprotocol.Choke:
		{
			t.logger.Info("torrent -> received CHOKE", "PeerID", string(pe.ID[:]))
			pe.AmChoked = true
		}
	case peerprotocol.Unchoke:
		{
			t.logger.Info("torrent -> received UNCHOKE", "PeerID", string(pe.ID[:]))
			pe.AmChoked = false

			picked, ok := t.picker.Pick(pe)
			if !ok {
				t.logger.Info("torrent -> couldn't pick piece for peer", "PeerID", string(pe.ID[:]))
				pe.IsInteresting = false
				return
			}

			downloader, ok := t.downloaders[pe.Peer]
			if !ok {
				pieceSize := t.meta.PieceLength
				if picked == uint32(t.meta.PieceCount)-1 {
					pieceSize = (t.meta.PieceLength * t.meta.PieceCount) - t.meta.Length
				}
				t.downloaders[pe.Peer] = piecedownloader.NewPieceDownloader(picked, uint32(pieceSize))
				downloader = t.downloaders[pe.Peer]
				downloader.Set(pe)
			}
			downloader.RequestBlocks(10)
		}
	case peerprotocol.Interested:
		{
			t.logger.Info("torrent -> received INTERESTED", "PeerID", string(pe.ID[:]))
			pe.AmInteresting = true
		}
	case peerprotocol.Uninterested:
		{
			t.logger.Info("torrent -> received UNINTERESTED", "PeerID", string(pe.ID[:]))
			pe.AmInteresting = false
		}
	case peerprotocol.Have:
		{
			if pe.Pieces == nil {
				pe.Pieces = bitset.MustNew(uint(t.meta.PieceCount))
			}
			if mess.Idx > uint32(t.meta.PieceCount-1) {
				t.logger.Info("torrent -> invalid HAVE", "PeerID", string(pe.ID[:]), "Error", "Piece index out of bounds")
				delete(t.downloaders, pe.Peer)
				pe.Stop()
				return
			}
			pe.Pieces.Set(uint(mess.Idx))
			t.picker.OnPeerHave(mess.Idx)
			if !t.pieces.Test(uint(mess.Idx)) {
				pe.IsInteresting = true
			}
			t.logger.Info("torrent -> received HAVE", "PeerID", string(pe.ID[:]), "Idx", mess.Idx)
		}
	case peerprotocol.Bitfield:
		{
			if pe.Pieces != nil {
				delete(t.downloaders, pe.Peer)
				pe.Stop()
				return
			}
			spareBits := len(mess.Pieces)*8 - int(t.meta.PieceCount)
			for i := range spareBits {
				if mess.Pieces[len(mess.Pieces)-1]&1<<i != 0 {
					t.logger.Info("torrent -> invalid BITFIELD", "PeerID", string(pe.ID[:]), "Error", "spare bits are set")
				}
			}
			data, err := util.BytesToBitset(mess.Pieces, uint(t.meta.PieceCount))
			if err != nil {
				t.logger.Info("torrent -> invalid BITFIELD", "PeerID", string(pe.ID[:]), "Error", err)
				delete(t.downloaders, pe.Peer)
				pe.Stop()
				return
			}

			for i := range data.EachSet() {
				if !t.pieces.Test(i) {
					pe.IsInteresting = true
					break
				}
			}

			pe.Pieces = data
			t.picker.OnPeerBitfield(pe)
			t.logger.Info("torrent -> received BITFIELD", "PeerID", string(pe.ID[:]), "Pieces", pe.Pieces.Count())
		}
	case peerprotocol.Request:
		t.logger.Info("torrent -> received REQUEST", "PeerID", string(pe.ID[:]), "Idx", mess.Idx, "Begin", mess.Begin, "Length", mess.Length)
	case peerprotocol.Piece:
		t.logger.Info("torrent -> received PIECE", "PeerID", string(pe.ID[:]), "Idx", mess.Idx, "Begin", mess.Begin, "Data Length", len(mess.Data))
		if pe.AmChoked {
			return
		}
		downloader, ok := t.downloaders[pe.Peer]
		if !ok {
			return
		}
		downloader.OnBlockReceived(mess.Begin, uint32(len(mess.Data)))
		downloader.RequestBlocks(10)
	case peerprotocol.Cancel:
		t.logger.Info("torrent -> received CANCEL", "PeerID", string(pe.ID[:]), "Idx", mess.Idx, "Begin", mess.Begin, "Length", mess.Length)
	}
}
