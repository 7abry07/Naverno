package sequentialpicker_test

import (
	"Naverno/internal/bitfield"
	"Naverno/internal/picker"
	"Naverno/internal/picker/sequentialpicker"
	"Naverno/internal/piece"
	"testing"
)

func TestPicker(t *testing.T) {
	pieces := make([]*piece.Piece, 20)
	for i := range len(pieces) {
		pieces[i] = piece.NewPiece(uint32(i), 10, 0, [20]byte{})
	}
	p := sequentialpicker.NewSequentialPicker(pieces)

	peerPieces := bitfield.New(20)
	peerPieces.Set(1).Set(10)

	pe := picker.NewMockPeer(peerPieces)

	piece1 := p.Pick(pe)
	if piece1 == nil {
		t.Fatal("Pick() failed")
	}
	if piece1.Idx != 1 {
		t.Errorf("expected index -> %v | got -> %v", 1, piece1)
	}

	piece2 := p.Pick(pe)
	if piece2 == nil {
		t.Fatal("Pick() failed")
	}
	if piece2.Idx != 10 {
		t.Errorf("expected index -> %v | got -> %v", 10, piece2)
	}
}

func TestPickerFail(t *testing.T) {
	pieces := make([]*piece.Piece, 20)
	for i := range len(pieces) {
		pieces[i] = piece.NewPiece(uint32(i), 10, 0, [20]byte{})
	}
	p := sequentialpicker.NewSequentialPicker(pieces)
	p.OnPieceCompleted(pieces[1])
	p.OnPieceCompleted(pieces[10])
	p.OnPieceCompleted(pieces[19])

	peerPieces := bitfield.New(20)
	peerPieces.Set(1).Set(10).Set(19)

	pe := picker.NewMockPeer(peerPieces)

	piece := p.Pick(pe)
	if piece != nil {
		t.Errorf("Pick() should have failed, got -> %v", piece)
	}
}
