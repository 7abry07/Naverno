package piecedownloader

import (
	"Naverno/internal/util"
	"fmt"
	"log/slog"
	"maps"
)

const (
	DefaultBlockSize = 1024 * 16
)

type PieceDownloader struct {
	logger    *slog.Logger
	Piece     uint32
	PieceSize uint32
	peer      Peer
	remaining map[uint32]uint32
	pending   map[uint32]uint32
}

func NewPieceDownloader(logger *slog.Logger, piece uint32, pieceSize uint32) *PieceDownloader {
	if logger == nil {
		panic("passed nil logger to piece downloader")
	}
	alignedPieceSize := util.Align(uint64(pieceSize), DefaultBlockSize)
	lastBlockSize := alignedPieceSize - uint64(pieceSize)
	if lastBlockSize == 0 {
		lastBlockSize = DefaultBlockSize
	}
	blocksPerPiece := alignedPieceSize / DefaultBlockSize
	blocks := make(map[uint32]uint32)

	for i := range blocksPerPiece {
		size := DefaultBlockSize
		if i == blocksPerPiece-1 {
			size = int(lastBlockSize)
		}
		blocks[uint32(i*DefaultBlockSize)] = uint32(size)
	}

	return &PieceDownloader{
		Piece:     piece,
		PieceSize: pieceSize,
		logger:    logger,
		peer:      nil,
		remaining: blocks,
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
		d.logger.Debug("downloader -> block requested", "Piece", d.Piece, "Block", fmt.Sprintf("(%v, %v)", begin, length))

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
	_, pending := d.pending[begin]
	if pending {
		delete(d.pending, begin)
	}
	delete(d.remaining, begin)
}
