package piecedownloader_test

import (
	"Naverno/internal/piecedownloader"
	"Naverno/internal/piecedownloader/downloadertest"
	"io"
	"log/slog"
	"testing"
)

func TestDownloader(t *testing.T) {
	d := piecedownloader.NewPieceDownloader(slog.New(slog.NewTextHandler(io.Discard, nil)), 5, piecedownloader.DefaultBlockSize*5)
	pe := downloadertest.NewMockPeer()

	d.Set(pe)
	d.RequestBlocks(3)

	pieces := pe.GetPieces()
	if len(pieces) != 3 {
		t.Errorf("requested blocks aren'tequal to queue size, pieces -> %v | queue size -> %v", len(pieces), 3)
	}
	for _, block := range pieces {
		err := d.OnBlockReceived(block.Begin, uint32(len(block.Data)))
		if err != nil {
			t.Errorf("unexpected error -> %v", err)
		}
	}

	d.RequestBlocks(2)
	pieces = pe.GetPieces()
	if len(pieces) != 2 {
		t.Errorf("requested blocks aren't equal to queue size, pieces -> %v | queue size -> %v", len(pieces), 2)
	}

	for _, block := range pieces {
		err := d.OnBlockReceived(block.Begin, uint32(len(block.Data)))
		if err != nil {
			t.Errorf("unexpected error -> %v", err)
		}
	}

	d.RequestBlocks(1)
	pieces = pe.GetPieces()
	if len(pieces) != 0 {
		t.Errorf("requested blocks on completed piece")
	}
	for _, block := range pieces {
		err := d.OnBlockReceived(block.Begin, uint32(len(block.Data)))
		if err != nil {
			t.Errorf("unexpected error -> %v", err)
		}
	}
}
