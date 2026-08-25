package piecedownloader_test

import (
	"Naverno/internal/peerprotocol"
	"Naverno/internal/piece"
	"Naverno/internal/piecedownloader"
	"io"
	"log/slog"
	"testing"
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

func (pe *MockPeer) Request(idx, begin, length uint32) {
	pe.requests = append(pe.requests, peerprotocol.Request{Idx: idx, Begin: begin, Length: length})
}

func TestDownloader(t *testing.T) {
	p := piece.NewPiece(5, piece.BlockSize*5, 0, [20]byte{})

	d := piecedownloader.New(slog.New(slog.NewTextHandler(io.Discard, nil)), p)
	pe := NewMockPeer()

	d.Set(pe)
	d.RequestBlocks(3)

	pieces := pe.GetPieces()
	if len(pieces) != 3 {
		t.Errorf("requested blocks aren'tequal to queue size, pieces -> %v | queue size -> %v", len(pieces), 3)
	}
	for _, block := range pieces {
		d.OnBlockReceived(block.Begin, uint32(len(block.Data)))
	}

	d.RequestBlocks(2)
	pieces = pe.GetPieces()
	if len(pieces) != 2 {
		t.Errorf("requested blocks aren't equal to queue size, pieces -> %v | queue size -> %v", len(pieces), 2)
	}

	for _, block := range pieces {
		d.OnBlockReceived(block.Begin, uint32(len(block.Data)))
	}

	d.RequestBlocks(1)
	pieces = pe.GetPieces()
	if len(pieces) != 0 {
		t.Errorf("requested blocks on completed piece")
	}
	for _, block := range pieces {
		d.OnBlockReceived(block.Begin, uint32(len(block.Data)))
	}
}
