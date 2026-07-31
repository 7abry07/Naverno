package rarestfirstpicker_test

import (
	"Naverno/internal/bitfield"
	"Naverno/internal/picker/rarestfirstpicker"
	"testing"
)

func TestRarestFirst(t *testing.T) {

	p := rarestfirstpicker.New(5)
	p.OnPeerHave(0)
	p.OnPeerHave(1)
	p.OnPeerHave(2)
	p.OnPeerHave(3)

	peerPieces := bitfield.New(5)
	peerPieces.SetAll()

	picked, ok := p.Pick(peerPieces)
	if !ok {
		t.Fatal("Pick() failed")
	}

	if picked != 4 {
		t.Errorf("should have picked piece with index %v, picked %v instead", 4, picked)
	}

	p.OnPeerDisconnected(peerPieces)
	picked, ok = p.Pick(peerPieces)
	if !ok {
		t.Fatal("Pick() failed")
	}
}
