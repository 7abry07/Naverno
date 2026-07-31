package sequentialpicker_test

import (
	"Naverno/internal/bitfield"
	"Naverno/internal/picker"
	"Naverno/internal/picker/sequentialpicker"
	"testing"
)

func TestPicker(t *testing.T) {
	p := sequentialpicker.New(20)

	peerPieces := bitfield.New(20)
	peerPieces.Set(1).Set(10)

	pe := picker.NewMockPeer(peerPieces)

	piece1, ok := p.Pick(pe)
	if !ok {
		t.Fatal("Pick() failed")
	}
	if piece1 != 1 {
		t.Errorf("expected index -> %v | got -> %v", 1, piece1)
	}

	piece2, ok := p.Pick(pe)
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

	pe := picker.NewMockPeer(peerPieces)

	piece, ok := p.Pick(pe)
	if ok {
		t.Errorf("Pick() should have failed, got -> %v", piece)
	}
}
