package piece_test

import (
	"Naverno/internal/piece"
	"testing"
)

func TestPieces(t *testing.T) {
	pieces := []*piece.Piece{}
	pieces = append(pieces, piece.NewPiece(0, piece.BlockSize*6, 0, [20]byte{}))
	pieces = append(pieces, piece.NewPiece(1, (piece.BlockSize*5)+1, 0, [20]byte{}))

	if pieces[0].BlockCount != 6 {
		t.Errorf("expected block count-> %v, got -> %v", 6, pieces[0].BlockCount)
	}
	if pieces[1].BlockCount != 6 {
		t.Errorf("expected block count-> %v, got -> %v", 6, pieces[1].BlockCount)
	}

	lastBlockBeginExpected := (pieces[1].BlockCount * piece.BlockSize) - piece.BlockSize
	lastBlockLengthExpected := 1
	lastBlockBegin := uint32(0)
	lastBlockLength := uint32(0)
	for begin, length := range pieces[1].Blocks() {
		lastBlockBegin = begin
		lastBlockLength = length
	}
	if lastBlockBegin != uint32(lastBlockBeginExpected) ||
		lastBlockLength != uint32(lastBlockLengthExpected) {
		t.Errorf("expected last block -> (%v, %v), got -> (%v, %v)", lastBlockBeginExpected, lastBlockLengthExpected, lastBlockBegin, lastBlockLength)
	}
}
