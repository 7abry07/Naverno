package downloadertest

import (
	"Naverno/internal/peerprotocol"
	"Naverno/internal/piece"
)

type MockPeer struct {
	requests []peerprotocol.Request
}

func NewMockPeer() *MockPeer {
	return &MockPeer{requests: []peerprotocol.Request{}}
}

func (pe *MockPeer) GetPieces() []peerprotocol.Piece {
	res := []peerprotocol.Piece{}
	for _, r := range pe.requests {
		res = append(res, peerprotocol.Piece{Idx: r.Idx, Begin: r.Begin, Data: make([]byte, r.Length)})
	}

	pe.requests = []peerprotocol.Request{}
	return res
}

func (pe *MockPeer) Request(pi *piece.Piece, begin, length uint32) bool {
	pe.requests = append(pe.requests, peerprotocol.Request{Idx: pi.Idx, Begin: begin, Length: length})
	return true
}
