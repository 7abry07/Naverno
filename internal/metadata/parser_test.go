package metadata_test

import (
	"GoBit/internal/metadata"
	"os"
	"testing"
)

func TestParseValidSingleTorrent(t *testing.T) {
	data, err := os.ReadFile("testdata/debian.torrent")
	if err != nil {
		t.Fatal(err)
	}

	meta, err := metadata.Parse(string(data))
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

func TestParseValidMultiTorrent(t *testing.T) {
	data, err := os.ReadFile("testdata/fedora.torrent")
	if err != nil {
		t.Fatal(err)
	}

	meta, err := metadata.Parse(string(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if meta.Info == nil {
		t.Fatal("expected info")
	}

	announce := "http://torrent.fedoraproject.org:6969/announce"
	if meta.Announce().String() != announce {
		t.Errorf("expected: [%v] | got: [%v]", announce, meta.Announce().String())
	}

	name := "Fedora-Budgie-Live-x86_64-44"
	if meta.Name() != name {
		t.Errorf("expected: [%v] | got: [%v]", name, meta.Name())
	}

	piece_length := 262144
	if meta.PieceLength() != piece_length {
		t.Errorf("expected: [%v] | got: [%v]", piece_length, meta.PieceLength())
	}

	pieces := 235340
	if len(meta.Pieces()) != pieces {
		t.Errorf("expected: [%v] | got: [%v]", pieces, len(meta.Pieces()))
	}

	switch info := meta.Info.(type) {
	case metadata.SingleFile:
		t.Fatalf("expected: [%v] | got: [%v]", "multi file torrent", "single file torrent")
	case metadata.MultiFile:
		path1 := "Fedora-Budgie-Live-44-1.7.x86_64.iso"
		path2 := "Fedora-Spins-44-1.7-x86_64-CHECKSUM"
		length1 := 3084500992
		length2 := 2922

		if info.Files()[0].Path != path1 {
			t.Fatalf("expected: [%v] | got: [%v]", path1, info.Files()[0].Path)
		}
		if info.Files()[1].Path != path2 {
			t.Fatalf("expected: [%v] | got: [%v]", path2, info.Files()[1].Path)
		}
		if info.Files()[0].Length != uint(length1) {
			t.Fatalf("expected: [%v] | got: [%v]", length1, info.Files()[0].Length)
		}

		if info.Files()[1].Length != uint(length2) {
			t.Fatalf("expected: [%v] | got: [%v]", length2, info.Files()[1].Length)
		}
	}
}
