package rarestfirstpicker_test

import (
	"Naverno/internal/bitfield"
	"Naverno/internal/picker"
	"Naverno/internal/picker/rarestfirstpicker"
	"Naverno/internal/piece"
	"testing"
)

func TestRarestFirst(t *testing.T) {

	pieces := make([]*piece.Piece, 20)
	for i := range len(pieces) {
		pieces[i] = piece.NewPiece(uint32(i), 10, 0, [20]byte{})
	}
	p := rarestfirstpicker.New(pieces)
	p.OnPeerHave(pieces[5])

	peerPieces := bitfield.New(20)
	peerPieces.SetAll()

	pe := picker.NewMockPeer(peerPieces)

	picked := p.Pick(pe)
	if picked.Idx != 5 {
		t.Errorf("should have picked piece with index %v, picked %v instead", 5, picked.Idx)
	}
}
