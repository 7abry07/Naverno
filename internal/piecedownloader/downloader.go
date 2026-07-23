package piecedownloader

import (
	"Naverno/internal/piece"
	"fmt"
	"log/slog"
	"maps"
)

type PieceDownloader struct {
	*piece.Piece
	peer      Peer
	logger    *slog.Logger
	remaining map[uint32]uint32
	pending   map[uint32]uint32
}

func NewPieceDownloader(logger *slog.Logger, p *piece.Piece) *PieceDownloader {
	if logger == nil {
		panic("passed nil logger to piece downloader")
	}

	return &PieceDownloader{
		Piece:     p,
		logger:    logger,
		peer:      nil,
		remaining: maps.Collect(p.Blocks()),
		pending:   make(map[uint32]uint32),
	}
}

func (d *PieceDownloader) Set(p Peer) {
	d.peer = p
}

func (d *PieceDownloader) RequestBlocks(queueSize int) {
	if d.peer == nil {
		panic("nil peer in downloader")
	}
	i := 1
	temp := []uint32{}
	for begin, length := range d.remaining {
		if len(d.pending) >= queueSize {
			break
		}

		d.peer.Request(d.Piece, begin, length)
		temp = append(temp, begin)
		d.pending[begin] = length
		d.logger.Debug("downloader -> block requested", "Piece", d.Piece.Idx, "Block", fmt.Sprintf("(%v, %v)", begin, length))

		i++
	}

	for _, begin := range temp {
		delete(d.remaining, begin)
	}

}

func (d *PieceDownloader) Completed() bool {
	return len(d.remaining) == 0 && len(d.pending) == 0
}

func (d *PieceDownloader) OnPeerDisconnected() {
	maps.Copy(d.remaining, d.pending)
	d.pending = make(map[uint32]uint32)
}

func (d *PieceDownloader) OnPeerChoke() {
	maps.Copy(d.remaining, d.pending)
	d.pending = make(map[uint32]uint32)
}

func (d *PieceDownloader) OnBlockReceived(begin uint32, length uint32) {
	delete(d.pending, begin)
	delete(d.remaining, begin)
}
