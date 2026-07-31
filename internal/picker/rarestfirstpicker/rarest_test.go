package rarestfirstpicker_test

import (
	"Naverno/internal/bitfield"
	"Naverno/internal/picker"
	"Naverno/internal/picker/rarestfirstpicker"
	"testing"
)

func TestRarestFirst(t *testing.T) {

	p := rarestfirstpicker.New(20)
	p.OnPeerHave(5)

	peerPieces := bitfield.New(20)
	peerPieces.SetAll()

	pe := picker.NewMockPeer(peerPieces)

	picked, ok := p.Pick(pe)
	if !ok {
		t.Fatal("Pick() failed")
	}

	if picked != 5 {
		t.Errorf("should have picked piece with index %v, picked %v instead", 5, picked)
	}
}
