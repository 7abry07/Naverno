package piecedownloader

import (
	"Naverno/internal/piece"
	"Naverno/internal/util"
	"errors"
	"fmt"
	"log/slog"
)

var (
	ErrInvalid      = errors.New("block invalid")
	ErrNotRequested = errors.New("block not requested")
	ErrDuplicate    = errors.New("block duplicate")
)

const (
	BlockSize = 16 * 1024
)

type PieceDownloader struct {
	*piece.Piece
	peer      Peer
	logger    *slog.Logger
	blocks    map[uint32]uint32
	remaining []uint32
	pending   map[uint32]struct{}
	done      map[uint32]struct{}
}

func New(logger *slog.Logger, p *piece.Piece) *PieceDownloader {
	if logger == nil {
		panic("passed nil logger to piece downloader")
	}
	d := &PieceDownloader{
		Piece:     p,
		logger:    logger,
		peer:      nil,
		blocks:    make(map[uint32]uint32),
		remaining: []uint32{},
		pending:   make(map[uint32]struct{}),
		done:      make(map[uint32]struct{}),
	}

	blockCount := uint32(util.Align(uint64(p.Size), BlockSize)) / BlockSize

	for i := range blockCount {
		length := uint64(BlockSize)
		if i == blockCount-1 {
			length -= util.Align(uint64(p.Size), BlockSize) - uint64(p.Size)
		}
		d.blocks[i*BlockSize] = uint32(length)
		d.remaining = append(d.remaining, i*BlockSize)
	}

	return d
}

func (d *PieceDownloader) Set(p Peer) {
	d.peer = p
}

func (d *PieceDownloader) RequestBlocks(queueSize int) {
	if d.peer == nil {
		panic("nil peer in downloader")
	}

	remaining := d.remaining
	for _, begin := range remaining {
		if len(d.pending) >= queueSize {
			break
		}
		length := d.blocks[begin]

		d.peer.Request(d.Piece.Idx, begin, length)
		d.pending[begin] = struct{}{}
		d.remaining = d.remaining[1:]
		d.logger.Debug("downloader -> block requested", "Piece", d.Piece.Idx, "Block", fmt.Sprintf("(%v, %v)", begin, length))
	}
}

func (d *PieceDownloader) Completed() bool {
	return len(d.done) == len(d.blocks)
}

func (d *PieceDownloader) OnPeerDisconnected() {
	for begin := range d.pending {
		d.remaining = append(d.remaining, begin)
	}
	d.pending = make(map[uint32]struct{})
}

func (d *PieceDownloader) OnPeerChoke() {
	for begin := range d.pending {
		d.remaining = append(d.remaining, begin)
	}
	d.pending = make(map[uint32]struct{})
}

func (d *PieceDownloader) OnBlockReceived(begin uint32, length uint32) error {
	if _, ok := d.blocks[begin]; !ok {
		return ErrInvalid
	}

	if _, ok := d.done[begin]; ok {
		return ErrDuplicate
	}

	if _, ok := d.pending[begin]; !ok {
		return ErrNotRequested
	}
	delete(d.pending, begin)
	d.done[begin] = struct{}{}
	return nil
}
