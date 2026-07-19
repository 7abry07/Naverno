package piecedownloader_test

import (
	"Naverno/internal/piecedownloader"
	"Naverno/internal/piecedownloader/downloadertest"
	"testing"
)

func TestDownloader(t *testing.T) {
	d := piecedownloader.NewPieceDownloader(5, piecedownloader.DefaultBlockSize*5)
	pe := downloadertest.NewMockPeer()

	d.Set(pe)
	ok := d.RequestBlocks(3)
	if !ok {
		t.Fatalf("RequestBlocks(%v) couldn't request said block count", 3)
	}

	pieces := pe.GetPieces()
	if len(pieces) != 3 {
		t.Errorf("requested blocks aren'tequal to queue size, pieces -> %v | queue size -> %v", len(pieces), 3)
	}

	ok = d.RequestBlocks(2)
	if !ok {
		t.Fatalf("RequestBlocks(%v) couldn't request said block count", 2)
	}
	pieces = pe.GetPieces()
	if len(pieces) != 2 {
		t.Errorf("requested blocks aren'tequal to queue size, pieces -> %v | queue size -> %v", len(pieces), 2)
	}
	if d.RequestBlocks(1) {
		t.Error("RequestBlocks() supposed to fail but didn't")
	}

	d = piecedownloader.NewPieceDownloader(1, piecedownloader.DefaultBlockSize*5)
	d.Set(pe)

	ok = d.RequestBlocks(3)
	if !ok {
		t.Fatalf("RequestBlocks(%v) couldn't request said block count", 3)
	}

	d.CancelPending()
	pieces = pe.GetPieces()
	if len(pieces) != 0 {
		t.Errorf("pending requests weren't canceled on CancelPending()")
	}
}
