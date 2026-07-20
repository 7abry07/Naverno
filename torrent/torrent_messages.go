package torrent

import (
	"Naverno/internal/peer"
	"Naverno/internal/peerprotocol"
	"Naverno/internal/piecedownloader"
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
			downloader.CancelPending()
		}
	case peerprotocol.Unchoke:
		{
			t.logger.Debug("torrent -> received UNCHOKE", "PeerID", string(pe.ID[:]))
			pe.AmChoked = false

			if pe.Pieces == nil {
				return
			}
			picked, ok := t.picker.Pick(pe)
			if !ok {
				pe.IsInteresting = false
				return
			}

			downloader, ok := t.downloaders[pe.Peer]
			if !ok {
				downloader, ok = t.stalledDownloaders[picked]
				if !ok {
					pieceSize := t.meta.PieceLength
					if picked == uint32(t.meta.PieceCount)-1 {
						pieceSize = (t.meta.PieceLength * t.meta.PieceCount) - t.meta.Length
					}
					t.downloaders[pe.Peer] = piecedownloader.NewPieceDownloader(t.logger, picked, uint32(pieceSize))
					downloader = t.downloaders[pe.Peer]
					downloader.Set(pe)
					t.logger.Info("torrent -> started downloader for piece", "Piece", picked, "PeerID", string(pe.ID[:]))
				} else {
					delete(t.stalledDownloaders, downloader.Piece)
					downloader.Set(pe)
					t.downloaders[pe.Peer] = downloader
					t.logger.Info("torrent -> restarted stalled downloader for piece", "Piece", downloader.Piece, "PeerID", string(pe.ID[:]))
				}
			}
			downloader.RequestBlocks(10)
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
				delete(t.downloaders, pe.Peer)
				pe.Stop()
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
			if !ok {
				downloader, ok = t.stalledDownloaders[mess.Idx]
				if !ok {
					pieceSize := t.meta.PieceLength
					if mess.Idx == uint32(t.meta.PieceCount)-1 {
						pieceSize = (t.meta.PieceLength * t.meta.PieceCount) - t.meta.Length
					}
					t.downloaders[pe.Peer] = piecedownloader.NewPieceDownloader(t.logger, mess.Idx, uint32(pieceSize))
					downloader = t.downloaders[pe.Peer]
					downloader.Set(pe)
					t.logger.Info("torrent -> started downloader for piece", "Piece", mess.Idx, "PeerID", string(pe.ID[:]))
				} else {
					delete(t.stalledDownloaders, downloader.Piece)
					downloader.Set(pe)
					t.downloaders[pe.Peer] = downloader
					t.logger.Info("torrent -> restarted stalled downloader for piece", "Piece", downloader.Piece, "PeerID", string(pe.ID[:]))
				}
			}
			downloader.RequestBlocks(10)
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
			t.logger.Debug("torrent -> received BITFIELD", "PeerID", string(pe.ID[:]), "Pieces", pe.Pieces.Count())
		}
	case peerprotocol.Request:
		t.logger.Debug("torrent -> received REQUEST", "PeerID", string(pe.ID[:]), "Request", fmt.Sprintf("%v, %v, %v", mess.Idx, mess.Begin, mess.Length))
	case peerprotocol.Piece:
		t.logger.Debug("torrent -> received PIECE", "PeerID", string(pe.ID[:]), "Block", fmt.Sprintf("%v, %v, %v", mess.Idx, mess.Begin, len(mess.Data)))
		if pe.AmChoked {
			return
		}
		downloader, ok := t.downloaders[pe.Peer]
		if !ok {
			return
		}
		err := downloader.OnBlockReceived(mess.Begin, uint32(len(mess.Data)))

		if err != nil {
			t.logger.Warn("torrent -> downloader error", "Piece", mess.Idx, "Error", err)
			delete(t.downloaders, pe.Peer)
			pe.Stop()
			return
		}
		downloader.RequestBlocks(10)

		if !downloader.Completed() {
			return
		}

		t.downloaded += int64(downloader.PieceSize)
		t.left = t.meta.Length - t.downloaded
		t.pieces.Set(uint(downloader.Piece))
		t.logger.Info("torrent -> piece completed", "Piece", mess.Idx, "Bytes Completed", t.downloaded, "Bytes Left", t.left)
		delete(t.downloaders, pe.Peer)

		picked, ok := t.picker.Pick(pe)
		if !ok {
			pe.IsInteresting = false
			return
		}

		downloader, ok = t.stalledDownloaders[picked]
		if !ok {
			pieceSize := t.meta.PieceLength
			if picked == uint32(t.meta.PieceCount)-1 {
				pieceSize = (t.meta.PieceLength * t.meta.PieceCount) - t.meta.Length
			}
			t.downloaders[pe.Peer] = piecedownloader.NewPieceDownloader(t.logger, picked, uint32(pieceSize))
			downloader = t.downloaders[pe.Peer]
			downloader.Set(pe)
			t.logger.Info("torrent -> started downloader for piece", "Piece", picked, "PeerID", string(pe.ID[:]))
		} else {
			delete(t.stalledDownloaders, downloader.Piece)
			downloader.Set(pe)
			t.downloaders[pe.Peer] = downloader
			t.logger.Info("torrent -> restarted downloader for piece", "Piece", downloader.Piece, "PeerID", string(pe.ID[:]))
		}
		downloader.RequestBlocks(10)
	case peerprotocol.Cancel:
		t.logger.Info("torrent -> received CANCEL", "PeerID", string(pe.ID[:]), "Idx", mess.Idx, "Begin", mess.Begin, "Length", mess.Length)
	}
}
