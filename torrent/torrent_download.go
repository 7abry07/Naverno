package torrent

import (
	"Naverno/internal/peer"
	"Naverno/internal/piece"
	"Naverno/internal/piecedownloader"
	"Naverno/internal/piecewriter"
)

func (t *Torrent) pieceCompleted(p *piece.Piece) {
	t.downloaded += int64(p.Size)
	t.left = t.meta.Length - t.downloaded
	t.bitset.Set(p.Idx)
	t.picker.OnPieceCompleted(p)
	t.logger.Info("torrent -> piece completed", "Piece", p.Idx, "Pieces Completed", t.bitset.Count())

	for pe := range t.peers {
		pe.Have(p.Idx)
	}
}

func (t *Torrent) handleWriterResult(res *piecewriter.PieceWriter) {
	if res.Err != nil {
		t.logger.Info("torrent -> error in piece writer", "Error", res.Err)
	}
}

func (t *Torrent) writePiece(p *piece.Piece, begin uint32, data []byte) {
	writer := piecewriter.New(p, begin, t.storage, data)
	t.writers[p] = writer
	go writer.Run(t.writerResults)
	t.logger.Debug("torrent -> started piece writer", "Piece", p.Idx, "Block", begin)
}

func (t *Torrent) download(pe *peer.Peer) {
	if pe.Pieces == nil || pe.AmChoked {
		return
	}

	downloader, ok := t.downloaders[pe]
	if ok {
		downloader.RequestBlocks(10)
		return
	}

	picked := t.picker.Pick(pe)
	if picked == nil {
		pe.IsInteresting = false
		return
	}

	downloader, ok = t.stalledDownloaders[picked]
	if ok {
		delete(t.stalledDownloaders, downloader.Piece)
		downloader.Set(pe)
		t.downloaders[pe] = downloader
		t.logger.Info("torrent -> restarted downloader for piece", "Piece", downloader.Piece.Idx, "PeerID", string(pe.ID[:]))
		downloader.RequestBlocks(10)
		return
	}
	t.downloaders[pe] = piecedownloader.NewPieceDownloader(t.logger, picked)
	downloader = t.downloaders[pe]
	downloader.Set(pe)
	downloader.RequestBlocks(10)
	t.logger.Debug("torrent -> started downloader for piece", "Piece", picked.Idx, "PeerID", string(pe.ID[:]))
}

func (t *Torrent) closeWriters() {
	for _, w := range t.writers {
		w.Close()
	}
}
