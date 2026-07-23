package piece

import (
	"Naverno/internal/metadata"
	"Naverno/internal/util"
	"cmp"
	"iter"
	"slices"
)

const (
	BlockSize = 1024 * 16
)

type Piece struct {
	Idx        uint32
	Offset     uint64
	Size       uint32
	BlockCount uint32
	Hash       [20]byte
}

func NewPieces(meta *metadata.Metadata) []*Piece {
	pieces := make([]*Piece, meta.PieceCount)

	off := uint64(0)
	for i := range pieces {
		size := meta.PieceLength
		if i == int(meta.PieceCount)-1 {
			size -= (meta.PieceLength * meta.PieceCount) - meta.Length
		}
		pieces[i] = NewPiece(uint32(i), uint32(size), off, [20]byte(meta.Pieces[i*20:(i*20)+20]))
		off += uint64(size)
	}

	slices.SortFunc(pieces, func(a, b *Piece) int { return cmp.Compare(a.Idx, b.Idx) })

	return pieces
}

func NewPiece(idx, size uint32, offset uint64, hash [20]byte) *Piece {
	return &Piece{
		Idx:        uint32(idx),
		Size:       uint32(size),
		Hash:       hash,
		Offset:     offset,
		BlockCount: uint32(util.Align(uint64(size), BlockSize)) / BlockSize,
	}
}

func (p *Piece) Blocks() iter.Seq2[uint32, uint32] {
	return func(yield func(uint32, uint32) bool) {
		for i := range p.BlockCount {
			length := uint64(BlockSize)
			if i == p.BlockCount-1 {
				length -= util.Align(uint64(p.Size), BlockSize) - uint64(p.Size)
			}
			if !yield(i*BlockSize, uint32(length)) {
				return
			}
		}
	}
}
