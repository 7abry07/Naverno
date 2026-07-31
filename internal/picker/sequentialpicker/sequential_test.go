package sequentialpicker_test

import (
	"Naverno/internal/bitfield"
	"Naverno/internal/picker/sequentialpicker"
	"testing"
)

func TestPicker(t *testing.T) {
	p := sequentialpicker.New(20)

	peerPieces := bitfield.New(20)
	peerPieces.Set(1).Set(10)

	piece1, ok := p.Pick(peerPieces)
	if !ok {
		t.Fatal("Pick() failed")
	}
	if piece1 != 1 {
		t.Errorf("expected index -> %v | got -> %v", 1, piece1)
	}

	piece2, ok := p.Pick(peerPieces)
	if !ok {
		t.Fatal("Pick() failed")
	}
	if piece2 != 10 {
		t.Errorf("expected index -> %v | got -> %v", 10, piece2)
	}
}

func TestPickerFail(t *testing.T) {
	p := sequentialpicker.New(20)
	p.OnPieceCompleted(1)
	p.OnPieceCompleted(10)
	p.OnPieceCompleted(19)

	peerPieces := bitfield.New(20)
	peerPieces.Set(1).Set(10).Set(19)

	piece, ok := p.Pick(peerPieces)
	if ok {
		t.Errorf("Pick() should have failed, got -> %v", piece)
	}
}
