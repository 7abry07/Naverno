package torrent

import (
	"Naverno/internal/peer"
	"Naverno/internal/peerprotocol"
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
				pe.Stop()
				return
			}
			pe.Pieces.Set(uint(mess.Idx))
			t.logger.Info("torrent -> received HAVE", "PeerID", string(pe.ID[:]), "Idx", mess.Idx)
		}
	case peerprotocol.Bitfield:
		{
			if pe.Pieces != nil {
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
				pe.Stop()
				return
			}

			pe.Pieces = data
			t.logger.Info("torrent -> received BITFIELD", "PeerID", string(pe.ID[:]))
		}
	case peerprotocol.Request:
		t.logger.Info("torrent -> received REQUEST", "PeerID", string(pe.ID[:]), "Idx", mess.Idx, "Begin", mess.Begin, "Length", mess.Length)
	case peerprotocol.Piece:
		t.logger.Info("torrent -> received PIECE", "PeerID", string(pe.ID[:]), "Idx", mess.Idx, "Begin", mess.Begin, "Data Length", len(mess.Data))
	case peerprotocol.Cancel:
		t.logger.Info("torrent -> received CANCEL", "PeerID", string(pe.ID[:]), "Idx", mess.Idx, "Begin", mess.Begin, "Length", mess.Length)
	}
}
