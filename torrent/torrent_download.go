package torrent

import (
	"Naverno/internal/peer"
	"Naverno/internal/piecedownloader"
)

func (t *Torrent) download(pe *peer.Peer, piece uint32) {
	downloader, ok := t.stalledDownloaders[piece]
	if ok {
		delete(t.stalledDownloaders, downloader.Piece)
		downloader.Set(pe)
		t.downloaders[pe] = downloader
		t.logger.Info("torrent -> restarted downloader for piece", "Piece", downloader.Piece, "PeerID", string(pe.ID[:]))
		downloader.RequestBlocks(10)
		return
	}
	pieceSize := t.meta.PieceLength
	if piece == uint32(t.meta.PieceCount)-1 {
		pieceSize -= (t.meta.PieceLength * t.meta.PieceCount) - t.meta.Length
	}
	t.downloaders[pe] = piecedownloader.NewPieceDownloader(t.logger, piece, uint32(pieceSize))
	downloader = t.downloaders[pe]
	downloader.Set(pe)
	downloader.RequestBlocks(10)
	t.logger.Debug("torrent -> started downloader for piece", "Piece", piece, "PeerID", string(pe.ID[:]))
}
