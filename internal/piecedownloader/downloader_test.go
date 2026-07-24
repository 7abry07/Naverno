package piecedownloader_test

import (
	"Naverno/internal/piece"
	"Naverno/internal/piecedownloader"
	"io"
	"log/slog"
	"testing"
)

func TestDownloader(t *testing.T) {
	p := piece.NewPiece(5, piece.BlockSize*5, 0, [20]byte{})

	d := piecedownloader.NewPieceDownloader(slog.New(slog.NewTextHandler(io.Discard, nil)), p)
	pe := piecedownloader.NewMockPeer()

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
