package piecedownloader

import (
	"Naverno/internal/peerprotocol"
	"Naverno/internal/util"
	"fmt"
)

const (
	DefaultBlockSize = 1024 * 16
)

type PieceDownloader struct {
	piece     uint32
	peer      Peer
	remaining map[uint32]uint32
	pending   map[peerprotocol.Request]struct{}
}

func NewPieceDownloader(piece uint32, pieceSize uint32) *PieceDownloader {
	alignedPieceSize := util.Align(uint64(pieceSize), DefaultBlockSize)
	lastBlockSize := alignedPieceSize - uint64(pieceSize)
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
		piece:     piece,
		peer:      nil,
		remaining: blocks,
		pending:   make(map[peerprotocol.Request]struct{}),
	}
}

func (d *PieceDownloader) Set(p Peer) {
	d.peer = p
}

func (d *PieceDownloader) RequestBlocks(queueSize uint32) {

	if d.peer == nil {
		panic("nil peer in downloader")
	}
	i := 0
	for begin, length := range d.remaining {
		request := peerprotocol.Request{Idx: d.piece, Begin: begin, Length: length}
		d.pending[request] = struct{}{}
		d.peer.Request(request.Idx, request.Begin, request.Length)
		delete(d.remaining, begin)

		if i == int(queueSize)-1 {
			return
		}
		i++
	}
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

func (d *PieceDownloader) OnBlockReceived(begin uint32, length uint32) error {
	_, ok := d.pending[peerprotocol.Request{Idx: d.piece, Begin: begin, Length: length}]
	if !ok {
		return fmt.Errorf("received piece that was not requested")
	}
	delete(d.pending, peerprotocol.Request{Idx: d.piece, Begin: begin, Length: length})

	return nil
}
