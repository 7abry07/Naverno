package torrent

import (
	"Naverno/internal/peer"
	"Naverno/internal/peerprotocol"
	// "Naverno/internal/piecedownloader"
	"Naverno/internal/util"
	"fmt"

	"github.com/bits-and-blooms/bitset"
)

func (t *Torrent) handlePeerMessage(pe peer.PeerMessage) {
	switch mess := pe.Message.(type) {
	case peerprotocol.Choke:
		{
			t.logger.Debug("torrent -> received CHOKE", "PeerID", string(pe.ID[:]))
			pe.AmChoked = true

			downloader, ok := t.downloaders[pe.Peer]
			if !ok {
				return
			}
			delete(t.downloaders, pe.Peer)
			t.stalledDownloaders[downloader.Piece] = downloader
			t.picker.OnPieceStalled(downloader.Piece)
			downloader.OnPeerChoke()
			t.logger.Info("torrent -> downloader stalled", "Piece", downloader.Piece)
		}
	case peerprotocol.Unchoke:
		{
			t.logger.Debug("torrent -> received UNCHOKE", "PeerID", string(pe.ID[:]))
			pe.AmChoked = false

			if pe.Pieces == nil {
				return
			}

			downloader, ok := t.downloaders[pe.Peer]
			if ok {
				downloader.RequestBlocks(10)
				return
			}

			picked, ok := t.picker.Pick(pe)
			if !ok {
				pe.IsInteresting = false
				return
			}
			t.download(pe.Peer, picked)
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
			if pe.Pieces == nil {
				pe.Pieces = bitset.MustNew(uint(t.meta.PieceCount))
			}
			if mess.Idx > uint32(t.meta.PieceCount-1) {
				t.logger.Info("torrent -> invalid HAVE", "PeerID", string(pe.ID[:]), "Error", "Piece index out of bounds")
				t.closePeer(pe.Peer)
				return
			}

			pe.Pieces.Set(uint(mess.Idx))
			t.picker.OnPeerHave(mess.Idx)
			if !t.pieces.Test(uint(mess.Idx)) {
				pe.IsInteresting = true
			}

			t.logger.Debug("torrent -> received HAVE", "PeerID", string(pe.ID[:]), "Idx", mess.Idx)

			if pe.AmChoked {
				return
			}

			downloader, ok := t.downloaders[pe.Peer]
			if ok {
				downloader.RequestBlocks(10)
				return
			}

			t.download(pe.Peer, mess.Idx)
		}
	case peerprotocol.Bitfield:
		{
			if pe.Pieces != nil {
				t.closePeer(pe.Peer)
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
				t.closePeer(pe.Peer)
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
			t.logger.Debug("torrent -> received BITFIELD", "PeerID", string(pe.ID[:]), "Pieces", pe.Pieces.Count())
		}
	case peerprotocol.Request:
		t.logger.Debug("torrent -> received REQUEST", "PeerID", string(pe.ID[:]), "Request", fmt.Sprintf("%v, %v, %v", mess.Idx, mess.Begin, mess.Length))
	case peerprotocol.Piece:
		{
			t.logger.Debug("torrent -> received PIECE", "PeerID", string(pe.ID[:]), "Block", fmt.Sprintf("%v, %v, %v", mess.Idx, mess.Begin, len(mess.Data)))
			if pe.AmChoked {
				return
			}
			downloader, ok := t.downloaders[pe.Peer]
			if !ok {
				return
			}
			downloader.OnBlockReceived(mess.Begin, uint32(len(mess.Data)))
			downloader.RequestBlocks(10)
			if !downloader.Completed() {
				return
			}

			t.downloaded += int64(downloader.PieceSize)
			t.left = t.meta.Length - t.downloaded
			t.pieces.Set(uint(downloader.Piece))
			t.picker.OnPieceCompleted(downloader.Piece)
			t.logger.Info("torrent -> piece completed", "Piece", mess.Idx, "Pieces Completed", t.pieces.Count())
			delete(t.downloaders, pe.Peer)

			for _, p := range t.peers {
				p.Have(downloader.Piece)
			}

			if t.pieces.All() {
				t.announceCompleted()
				t.logger.Info("torrent -> completed")
				return
			}

			picked, ok := t.picker.Pick(pe)
			if !ok {
				pe.IsInteresting = false
				return
			}
			t.download(pe.Peer, picked)
		}
	case peerprotocol.Cancel:
		t.logger.Info("torrent -> received CANCEL", "PeerID", string(pe.ID[:]), "Idx", mess.Idx, "Begin", mess.Begin, "Length", mess.Length)
	}
}
