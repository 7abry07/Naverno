package piecedownloader

import (
	"Naverno/internal/peerprotocol"
	"Naverno/internal/util"
	"fmt"
	"log/slog"
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
	pending   map[peerprotocol.Request]struct{}
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
		pending:   make(map[peerprotocol.Request]struct{}),
	}

}

func (d *PieceDownloader) Set(p Peer) {
	d.peer = p
}

func (d *PieceDownloader) RequestBlocks(queueSize int) {
	if d.peer == nil {
		panic("nil peer in downloader")
	}
	i := 0
	blocksRequest := []uint32{}
	for begin, length := range d.remaining {
		if i == queueSize-1 {
			d.logger.Debug("downloader -> requested blocks", "Piece", d.Piece, "Blocks Requested", queueSize, "Remaining Blocks", len(d.remaining))
			break
		}
		if len(d.pending) >= queueSize {
			d.logger.Debug("downloader -> max pending requests reached", "Piece", d.Piece, "Remaining Blocks", len(d.remaining))
			break
		}
		d.pending[peerprotocol.Request{Idx: d.Piece, Begin: begin, Length: length}] = struct{}{}

		fmt.Printf("requested -> (%v, %#v, %#v)\n", d.Piece, begin, length)

		d.peer.Request(d.Piece, begin, length)
		blocksRequest = append(blocksRequest, begin)

		i++
	}

	for _, begin := range blocksRequest {
		delete(d.remaining, begin)
	}

	d.logger.Debug("downloader -> all blocks requested", "Piece", d.Piece, "Pending", len(d.pending))
}

func (d *PieceDownloader) CancelPending() {
	if d.peer == nil {
		panic("nil peer in downloader")
	}
	for pending := range d.pending {
		d.peer.Cancel(pending.Idx, pending.Begin, pending.Length)
		d.remaining[pending.Begin] = pending.Length
	}
	d.pending = make(map[peerprotocol.Request]struct{})
}

func (d *PieceDownloader) Completed() bool {
	return len(d.remaining) == 0 && len(d.pending) == 0
}

func (d *PieceDownloader) OnPeerDisconnected() {
	for pending := range d.pending {
		d.remaining[pending.Begin] = pending.Length
	}
	d.pending = make(map[peerprotocol.Request]struct{})
}

func (d *PieceDownloader) OnBlockReceived(begin uint32, length uint32) error {
	_, ok := d.pending[peerprotocol.Request{Idx: d.Piece, Begin: begin, Length: length}]
	if !ok {
		fmt.Printf("pending -> %v\n", d.pending)
		return fmt.Errorf("received piece that was not requested (%v, %v)", begin, length)
	}
	delete(d.pending, peerprotocol.Request{Idx: d.Piece, Begin: begin, Length: length})

	return nil
}
