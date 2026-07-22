package pickertest

import (
	"Naverno/internal/bitfield"
)

type MockPeer struct {
	pieces *bitfield.Bitfield
}

func NewMockPeer(b *bitfield.Bitfield) *MockPeer {
	return &MockPeer{pieces: b}
}

func (pe *MockPeer) GetPieces() *bitfield.Bitfield {
	return pe.pieces
}
