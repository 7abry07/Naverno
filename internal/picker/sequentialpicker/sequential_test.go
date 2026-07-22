package sequentialpicker_test

import (
	"Naverno/internal/bitfield"
	"Naverno/internal/picker/pickertest"
	"Naverno/internal/picker/sequentialpicker"
	"testing"
)

func TestPicker(t *testing.T) {
	picker := sequentialpicker.NewSequentialPicker(20)

	peerPieces := bitfield.New(20)
	peerPieces.Set(1).Set(10)

	pe := pickertest.NewMockPeer(peerPieces)

	piece1, ok := picker.Pick(pe)
	if !ok {
		t.Fatal("Pick() failed")
	}
	if piece1 != 1 {
		t.Errorf("expected index -> %v | got -> %v", 1, piece1)
	}

	piece2, ok := picker.Pick(pe)
	if !ok {
		t.Fatal("Pick() failed")
	}
	if piece2 != 10 {
		t.Errorf("expected index -> %v | got -> %v", 10, piece2)
	}
}

func TestPickerFail(t *testing.T) {
	picker := sequentialpicker.NewSequentialPicker(20)
	picker.OnPieceCompleted(1)
	picker.OnPieceCompleted(10)
	picker.OnPieceCompleted(19)

	peerPieces := bitfield.New(20)
	peerPieces.Set(1).Set(10).Set(19)

	pe := pickertest.NewMockPeer(peerPieces)

	piece, ok := picker.Pick(pe)
	if ok {
		t.Errorf("Pick() should have failed, got -> %v", piece)
	}
}
