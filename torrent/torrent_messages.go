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
			t.logger.Info("torrent -> received CHOKE message", "PeerID", string(pe.ID[:]))
			pe.AmChoked = true
		}
	case peerprotocol.Unchoke:
		{
			t.logger.Info("torrent -> received UNCHOKE message", "PeerID", string(pe.ID[:]))
			pe.AmChoked = false
		}
	case peerprotocol.Interested:
		{
			t.logger.Info("torrent -> received INTERESTED message", "PeerID", string(pe.ID[:]))
			pe.AmInteresting = true
		}
	case peerprotocol.Uninterested:
		{
			t.logger.Info("torrent -> received UNINTERESTED message", "PeerID", string(pe.ID[:]))
			pe.AmInteresting = false
		}
	case peerprotocol.Have:
		{
			if pe.Pieces == nil {
				pe.Pieces = bitset.MustNew(uint(t.meta.PieceCount))
			}

			if mess.Idx > uint32(t.meta.PieceCount-1) {
				t.logger.Info("torrent -> invalid HAVE message", "PeerID", string(pe.ID[:]), "Error", "Piece index out of bounds")
				pe.Stop()
				return
			}
			pe.Pieces.Set(uint(mess.Idx))
			t.logger.Info("torrent -> received HAVE message", "PeerID", string(pe.ID[:]), "Idx", mess.Idx)
		}
	case peerprotocol.Bitfield:
		{
			if pe.Pieces == nil {
				minimumBits := ((t.meta.PieceCount + 7) / 8) * 8
				if len(mess.Pieces)*8 != int(minimumBits) {
					t.logger.Info("torrent -> invalid BITFIELD message", "PeerID", string(pe.ID[:]), "Error", "Invalid length")
					pe.Stop()
					return
				}
				data := make([]byte, minimumBits/8)
				copy(data, mess.Pieces)
				pe.Pieces = bitset.FromWithLength(uint(t.meta.PieceCount), util.BytesToUint64s(data))
				t.logger.Info("torrent -> received BITFIELD message", "PeerID", string(pe.ID[:]))
				return
			}
			pe.Stop()
		}
	case peerprotocol.Request:
		t.logger.Info("torrent -> received REQUEST message", "PeerID", string(pe.ID[:]), "Idx", mess.Idx, "Begin", mess.Begin, "Length", mess.Length)
	case peerprotocol.Piece:
		t.logger.Info("torrent -> received PIECE message", "PeerID", string(pe.ID[:]), "Idx", mess.Idx, "Begin", mess.Begin, "Data Length", len(mess.Data))
	case peerprotocol.Cancel:
		t.logger.Info("torrent -> received CANCEL message", "PeerID", string(pe.ID[:]), "Idx", mess.Idx, "Begin", mess.Begin, "Length", mess.Length)
	}
}
