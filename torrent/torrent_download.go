package torrent

import (
	"Naverno/internal/bitfield"
	"Naverno/internal/choker"
	"Naverno/internal/infodownloader"
	"Naverno/internal/metadata"
	"Naverno/internal/peer"
	"Naverno/internal/picker"
	"Naverno/internal/piece"
	"Naverno/internal/piecedownloader"
	"Naverno/internal/storage"
	"Naverno/internal/storage/posixstorage"
	"bytes"
	"crypto/sha1"
	"fmt"
	"time"
)

func (t *Torrent) handleHashResult(res storage.HashResult) {
	if res.Err != nil {
		t.err = fmt.Errorf("error while checking hash -> %v", res.Err)
		t.logger.Error("torrent -> error while checking hash", "Error", res.Err)
		return
	}
	if !res.Ok {
		t.logger.Warn("torrent -> hash doesn't match", "Piece", res.Piece.Idx)
		t.picker.SetFree(res.Piece.Idx)
		return
	}

	t.downloaded += int64(res.Piece.Size)
	t.left = t.info.Length - t.downloaded
	t.bitset.Set(uint(res.Piece.Idx))
	t.picker.OnPieceCompleted(res.Piece.Idx)
	for _, pe := range t.peers {
		pe.Have(res.Piece.Idx)
	}
	t.logger.Info("torrent -> piece completed", "Piece", res.Piece.Idx, "Completed", t.bitset.Count())

	if t.bitset.All() {
		t.closeSeeds()
		t.announceCompleted()
		t.choker.Seeding()
		t.logger.Info("torrent -> completed")
		return
	}
}

func (t *Torrent) handleWriteResult(res storage.WriteResult) {
	if res.Err != nil {
		t.err = fmt.Errorf("error while writing piece-> %v", res.Err)
		t.logger.Error("torrent -> error while writing piece", "Error", res.Err)
		return
	}
}

func (t *Torrent) pieceCompleted(p *piece.Piece) {
	t.storage.AsyncHash(t.hashResults, p)
	t.logger.Debug("torrent -> started hash checker", "Piece", p.Idx)
}

func (t *Torrent) writePiece(p *piece.Piece, begin uint32, data []byte) {
	t.storage.AsyncWrite(t.writeResults, p, begin, data)
	t.logger.Debug("torrent -> started piece writer", "Piece", p.Idx, "Block", begin)
}

func (t *Torrent) download(pe *peer.Peer) {
	if t.err != nil {
		return
	}

	if pe.Pieces == nil || pe.AmChoked {
		return
	}

	downloader, ok := t.downloaders[pe]
	if ok {
		downloader.RequestBlocks(10)
		return
	}

	var picked uint32
	switch t.pickerStrategy {
	case SEQUENTIAL_PIECE_SELECTION:
		picked, ok = t.picker.Pick(picker.SEQUENTIAL, pe.Pieces)
	case RAREST_FIRST_PIECE_SELECTION:
		picked, ok = t.picker.Pick(picker.RAREST_FIRST, pe.Pieces)
	}

	if !ok {
		pe.IsInteresting = false
		return
	}

	downloader, ok = t.stalledDownloaders[t.pieces[picked]]
	if ok {
		delete(t.stalledDownloaders, downloader.Piece)
		downloader.Set(pe)
		t.downloaders[pe] = downloader
		t.logger.Info("torrent -> restarted downloader for piece", "Piece", downloader.Piece.Idx, "PeerID", string(pe.ID[:]))
		downloader.RequestBlocks(10)
		return
	}
	t.downloaders[pe] = piecedownloader.New(t.logger, t.pieces[picked])
	downloader = t.downloaders[pe]
	downloader.Set(pe)
	downloader.RequestBlocks(10)
	t.logger.Debug("torrent -> started downloader for piece", "Piece", picked, "PeerID", string(pe.ID[:]))
}

func (t *Torrent) downloadMetadata(pe *peer.Peer) {
	if t.infoDownloader == nil {
		if pe.ExtendedHS.MetadataSize != 0 {
			t.infoDownloader = infodownloader.New(pe.ExtendedHS.MetadataSize)
			t.logger.Info("torrent -> downloading metadata")
		}
	}
	t.infoDownloader.AddPeer(pe)
	t.infoDownloader.Request(10)
}

func (t *Torrent) metadataCompleted(data []byte) bool {
	hasher := sha1.New()
	hasher.Write(data)
	hash := [20]byte(hasher.Sum(nil))

	if !bytes.Equal(hash[:], t.infohash[:]) {
		t.logger.Warn("torrent -> info hash check failed")
		t.infoDownloader.Reset()
		return false
	}

	info, err := metadata.NewInfo(data)
	if err != nil {
		t.logger.Error("torrent -> unexpected error in torrent info creation", "Error", err)
		t.infoDownloader.Reset()
		return false
	}

	t.infoDownloader = nil

	t.logger.Info("torrent -> metadata completed")
	t.info = info

	t.picker = picker.New(uint32(t.info.PieceCount))
	t.choker = choker.New(time.Second*10, time.Second*30)
	go t.choker.Run(t.chokerEvents)
	t.pieces = piece.NewPieces(info)
	t.bitset = bitfield.New(uint32(info.PieceCount))
	t.storage = posixstorage.New(t.logger, info.Files, t.savePath)
	t.left = info.Length
	return true
}
