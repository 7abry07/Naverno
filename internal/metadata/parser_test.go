package metadata_test

import (
	"GoBit/internal/metadata"
	"os"
	"testing"
)

func loadTorrent(t *testing.T) string {
	data, err := os.ReadFile("testdata/debian.torrent")
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestParseValidSingleTorrent(t *testing.T) {
	meta, err := metadata.Parse(loadTorrent(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if meta.Info == nil {
		t.Fatal("expected info")
	}

	announce := "http://bttracker.debian.org:6969/announce"
	if meta.Announce().String() != announce {
		t.Errorf("expected: [%v] | got: [%v]", announce, meta.Announce().String())
	}

	name := "debian-13.5.0-amd64-netinst.iso"
	if meta.Name() != name {
		t.Errorf("expected: [%v] | got: [%v]", name, meta.Name())
	}

	piece_length := 262144
	if meta.PieceLength() != piece_length {
		t.Errorf("expected: [%v] | got: [%v]", piece_length, meta.PieceLength())
	}

	pieces := 60400
	if len(meta.Pieces()) != pieces {
		t.Errorf("expected: [%v] | got: [%v]", pieces, len(meta.Pieces()))
	}

	switch info := meta.Info.(type) {
	case metadata.SingleFile:
		length := 791674880
		if info.Length() != length {
			t.Errorf("expected: [%v] | got: [%v]", length, info.Length())
		}
	case metadata.MultiFile:
		t.Fatalf("expected: [%v] | got: [%v]", "single file torrent", "multi file torrent")
	}
}
