package torrent

import (
	"Naverno/internal/bitfield"
	"Naverno/internal/hashchecker"
	"Naverno/internal/peer"
	"Naverno/internal/piece"
	"Naverno/internal/piecedownloader"
	"Naverno/internal/piecewriter"
)

func (t *Torrent) handleHasherResult(res *hashchecker.HashChecker) {
	if res.Err != nil {
		t.logger.Error("torrent -> error in piece writer", "Error", res.Err)
		t.session.RemoveTorrent(t)
		return
	}
	if !res.Matches {
		t.logger.Warn("torrent -> hash doesn't match", "Piece", res.Piece.Idx)
		t.picker.SetFree(res.Piece)
		return
	}

	t.downloaded += int64(res.Piece.Size)
	t.left = t.meta.Length - t.downloaded
	t.bitset.Set(res.Piece.Idx)
	t.picker.OnPieceCompleted(res.Piece)
	for pe := range t.peers {
		pe.Have(res.Piece.Idx)
	}
	t.logger.Info("torrent -> piece completed", "Piece", res.Piece.Idx, "Pieces Completed", t.bitset.Count())

	if t.bitset.All() {
		t.closeSeeds()
		t.announceCompleted()
		t.logger.Info("torrent -> completed")
		return
	}
}

func (t *Torrent) handleWriterResult(res *piecewriter.PieceWriter) {
	if res.Err != nil {
		t.logger.Error("torrent -> error in piece writer", "Error", res.Err)
		t.session.RemoveTorrent(t)
		return
	}
}

func (t *Torrent) pieceCompleted(p *piece.Piece) {
	hasher := hashchecker.New(t.storage, p)
	t.hashers[p] = hasher
	go hasher.Run(t.hashersResults)
	t.logger.Debug("torrent -> started hash checker", "Piece", p.Idx)
}

func (t *Torrent) writePiece(p *piece.Piece, begin uint32, data []byte) {
	writer := piecewriter.New(p, begin, t.storage, data)
	t.writers[p] = writer
	go writer.Run(t.writersResults)
	t.logger.Debug("torrent -> started piece writer", "Piece", p.Idx, "Block", begin)
}

func (t *Torrent) download(pe *peer.Peer) {
	if (pe.Pieces == bitfield.Bitfield{} || pe.AmChoked) {
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

func (t *Torrent) closeHashers() {
	for _, c := range t.hashers {
		c.Close()
	}
}
