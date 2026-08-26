package torrent

import (
	"Naverno/internal/bitfield"
	"Naverno/internal/infodownloader"
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
			if t.info == nil {
				return
			}

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
			if t.info == nil {
				return
			}
			t.download(pe.Peer)
		}
	case peerprotocol.Interested:
		{
			pe.AmInteresting = true
			if t.info == nil {
				return
			}

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
			if t.info == nil {
				return
			}

			if pe.Pieces == nil {
				pe.Pieces = bitfield.New(uint32(t.info.PieceCount))
			}
			if mess.Idx > uint32(t.info.PieceCount-1) {
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
			if t.info == nil {
				return
			}

			if pe.Pieces != nil {
				t.closePeer(pe.Peer)
				return
			}
			data, err := bitfield.From(mess.Pieces, uint32(t.info.PieceCount))
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

			t.downloadedSince += uint64(len(mess.Data))
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
	case peerprotocol.Extended:
		{
			if !pe.SupportsExtensionProtocol() {
				t.closePeer(pe.Peer)
				return
			}

			switch extended := mess.ExtendedMessage.(type) {
			case peerprotocol.ExtendedHandshake:
				pe.ExtendedHS = &extended
				t.logger.Info("torrent -> EXTENDED handshake", "Peer", string(pe.ID[:]))
				if t.info == nil && pe.SupportsUTMetadata() {
					t.downloadMetadata(pe.Peer)
				}
			case peerprotocol.UTMetadataRequest:
				t.logger.Info("torrent -> EXTENDED metadata request", "Peer", string(pe.ID[:]), "Piece", extended.Piece)
				if t.info != nil {
					pieceLen := min(len(t.info.Raw[(16*1024)*extended.Piece:]), 16*1024)
					pe.SendMetadata(extended.Piece, t.info.Raw[(16*1024)*extended.Piece:pieceLen])
				} else {
					pe.RejectMetadataRequest(extended.Piece)
				}
			case peerprotocol.UTMetadataResponse:
				t.logger.Info("torrent -> EXTENDED metadata response", "Peer", string(pe.ID[:]), "Piece", extended.Piece, "Length", len(extended.Data))
				if t.infoDownloader == nil {
					return
				}
				err := t.infoDownloader.OnPiece(extended.Piece, extended.Data)
				switch err {
				case infodownloader.ErrInvalid:
					t.logger.Info("downloader -> invalid metadata piece", "PeerID", string(pe.ID[:]), "Piece", extended.Piece, "Length", len(extended.Data))
					t.closePeer(pe.Peer)
					return
				case infodownloader.ErrDuplicate:
					t.logger.Info("downloader -> duplicate metadata piece", "PeerID", string(pe.ID[:]), "Piece", extended.Piece)
					return
				case infodownloader.ErrNotRequested:
					t.logger.Info("downloader -> not requested metadata piece", "PeerID", string(pe.ID[:]), "Piece", extended.Piece)
				case nil:
				default:
					t.closePeer(pe.Peer)
					return
				}
				data, completed := t.infoDownloader.Completed()
				if completed {
					if t.metadataCompleted(data) {
						return
					}
				}
				t.downloadMetadata(pe.Peer)
			case peerprotocol.UTMetadataReject:
				t.logger.Info("torrent -> EXTENDED metadata reject", "Peer", string(pe.ID[:]), "piece", extended.Piece)
				if t.infoDownloader == nil {
					return
				}
				err := t.infoDownloader.OnReject(extended.Piece)
				switch err {
				case infodownloader.ErrInvalid:
					t.logger.Info("downloader -> invalid metadata piece", "PeerID", string(pe.ID[:]), "Piece", extended.Piece)
					t.closePeer(pe.Peer)
					return
				case infodownloader.ErrNotRequested:
					t.logger.Info("downloader -> not requested metadata piece", "PeerID", string(pe.ID[:]), "Piece", extended.Piece)
				case nil:
				default:
					t.closePeer(pe.Peer)
					return
				}
				t.downloadMetadata(pe.Peer)
			}
		}
	}
}
