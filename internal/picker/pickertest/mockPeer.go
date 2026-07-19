package pickertest

import "github.com/bits-and-blooms/bitset"

type MockPeer struct {
	pieces *bitset.BitSet
}

func NewMockPeer(b *bitset.BitSet) *MockPeer {
	return &MockPeer{pieces: b}
}

func (pe *MockPeer) GetPieces() *bitset.BitSet {
	return pe.pieces
}
