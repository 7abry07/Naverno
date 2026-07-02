package metadata_test

import (
	"Naverno/internal/metadata"
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

	files := meta.Files()
	if len(files) != 1 {
		t.Fatalf("expected: [%v files] | got: [%v files]", 1, len(files))
	}

	length := 791674880
	path := name

	if files[0].Length != length {
		t.Errorf("expected: [%v] | got: [%v]", length, files[0].Length)
	}

	if files[0].Path != path {
		t.Errorf("expected: [%v] | got: [%v]", path, files[0].Path)
	}

	infohash := [20]byte{
		0x58, 0x84, 0x68, 0x60,
		0xf0, 0xa7, 0x66, 0xf8,
		0xa4, 0x2b, 0x0b, 0xb2,
		0x14, 0xd8, 0xc7, 0x13,
		0xfd, 0xf1, 0xb1, 0x67,
	}

	if meta.Infohash() != infohash {
		t.Errorf("expected: [%v] | got: [%v]", meta.Infohash(), infohash)
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

	files := meta.Files()
	if len(files) != 2 {
		t.Fatalf("expected: [%v files] | got: [%v files]", 2, len(files))
	}

	path1 := "Fedora-Budgie-Live-44-1.7.x86_64.iso"
	path2 := "Fedora-Spins-44-1.7-x86_64-CHECKSUM"
	length1 := 3084500992
	length2 := 2922

	if files[0].Path != path1 {
		t.Fatalf("expected: [%v] | got: [%v]", path1, files[0].Path)
	}
	if files[1].Path != path2 {
		t.Fatalf("expected: [%v] | got: [%v]", path2, files[1].Path)
	}
	if files[0].Length != length1 {
		t.Fatalf("expected: [%v] | got: [%v]", length1, files[0].Length)
	}

	if files[1].Length != length2 {
		t.Fatalf("expected: [%v] | got: [%v]", length2, files[1].Length)
	}

	infohash := [20]byte{
		0x78, 0xbe, 0x65, 0x13,
		0xb4, 0x49, 0xf8, 0x18,
		0x82, 0xc8, 0x52, 0x07,
		0x5c, 0xf0, 0x11, 0x61,
		0xf3, 0xb7, 0xb2, 0xa3,
	}

	if meta.Infohash() != infohash {
		t.Errorf("expected: [%v] | got: [%v]", meta.Infohash(), infohash)
	}
}
