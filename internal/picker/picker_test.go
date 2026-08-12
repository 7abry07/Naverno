package picker_test

import (
	"Naverno/internal/bitfield"
	"Naverno/internal/picker"
	"testing"
)

func TestRarestFirst(t *testing.T) {
	p := picker.New(5)
	p.OnPeerHave(0)
	p.OnPeerHave(1)
	p.OnPeerHave(2)
	p.OnPeerHave(3)

	peerPieces := bitfield.New(5)
	peerPieces.SetAll()

	picked, ok := p.Pick(picker.RAREST_FIRST, peerPieces)
	if !ok {
		t.Fatal("Pick() failed")
	}

	if picked != 4 {
		t.Errorf("should have picked piece with index %v, picked %v instead", 4, picked)
	}

	picked, ok = p.Pick(picker.RAREST_FIRST, peerPieces)
	if !ok {
		t.Fatal("Pick() failed")
	}

	if picked == 4 {
		t.Errorf("shouldn't have picked piece already downloading piece")
	}

	p.OnPeerDisconnected(peerPieces)
	picked, ok = p.Pick(picker.RAREST_FIRST, peerPieces)
	if !ok {
		t.Fatal("Pick() failed")
	}

	p = picker.New(5)
	peerPieces.ClearAll()
	picked, ok = p.Pick(picker.RAREST_FIRST, peerPieces)
	if ok {
		t.Errorf("Pick() should have failed, got -> %v", picked)
	}
}

func TestSequential(t *testing.T) {
	p := picker.New(20)

	peerPieces := bitfield.New(20)
	peerPieces.Set(1).Set(10)

	piece1, ok := p.Pick(picker.SEQUENTIAL, peerPieces)
	if !ok {
		t.Fatal("Pick() failed")
	}
	if piece1 != 1 {
		t.Errorf("expected index -> %v | got -> %v", 1, piece1)
	}

	piece2, ok := p.Pick(picker.SEQUENTIAL, peerPieces)
	if !ok {
		t.Fatal("Pick() failed")
	}
	if piece2 != 10 {
		t.Errorf("expected index -> %v | got -> %v", 10, piece2)
	}

	p = picker.New(20)
	p.OnPieceCompleted(1)
	p.OnPieceCompleted(10)
	p.OnPieceCompleted(19)

	peerPieces = bitfield.New(20)
	peerPieces.Set(1).Set(10).Set(19)

	piece, ok := p.Pick(picker.SEQUENTIAL, peerPieces)
	if ok {
		t.Errorf("Pick() should have failed, got -> %v", piece)
	}
}
