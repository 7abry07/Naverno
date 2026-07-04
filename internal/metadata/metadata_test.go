package metadata_test

import (
	"Naverno/internal/metadata"
	"os"
	"testing"
)

func TestParseValidSingleTorrent(t *testing.T) {
	file, err := os.Open("testdata/debian.torrent")
	if err != nil {
		t.Fatal(err)
	}

	meta, err := metadata.New(file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	announce := "http://bttracker.debian.org:6969/announce"
	if meta.AnnounceList[0][0].String() != announce {
		t.Errorf("expected: [%v] | got: [%v]", announce, meta.AnnounceList[0][0].String())
	}

	name := "debian-13.5.0-amd64-netinst.iso"
	if meta.Name != name {
		t.Errorf("expected: [%v] | got: [%v]", name, meta.Name)
	}

	piece_length := int64(262144)
	if meta.PieceLength != piece_length {
		t.Errorf("expected: [%v] | got: [%v]", piece_length, meta.PieceLength)
	}

	pieces := 60400
	if len(meta.Pieces) != pieces {
		t.Errorf("expected: [%v] | got: [%v]", pieces, len(meta.Pieces))
	}

	if len(meta.Files) != 1 {
		t.Fatalf("expected: [%v files] | got: [%v files]", 1, len(meta.Files))
	}

	length := int64(791674880)
	path := name

	if meta.Files[0].Length != length {
		t.Errorf("expected: [%v] | got: [%v]", length, meta.Files[0].Length)
	}

	if meta.Files[0].Path != path {
		t.Errorf("expected: [%v] | got: [%v]", path, meta.Files[0].Path)
	}

	infohash := [20]byte{
		0x58, 0x84, 0x68, 0x60,
		0xf0, 0xa7, 0x66, 0xf8,
		0xa4, 0x2b, 0x0b, 0xb2,
		0x14, 0xd8, 0xc7, 0x13,
		0xfd, 0xf1, 0xb1, 0x67,
	}

	if meta.Infohash != infohash {
		t.Errorf("expected: [%x] | got: [%x]", infohash, meta.Infohash)
	}
}

func TestParseValidMultiTorrent(t *testing.T) {
	file, err := os.Open("testdata/fedora.torrent")
	if err != nil {
		t.Fatal(err)
	}

	meta, err := metadata.New(file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	announce := "http://torrent.fedoraproject.org:6969/announce"
	if meta.AnnounceList[0][0].String() != announce {
		t.Errorf("expected: [%v] | got: [%v]", announce, meta.AnnounceList[0][0].String())
	}

	name := "Fedora-Budgie-Live-x86_64-44"
	if meta.Name != name {
		t.Errorf("expected: [%v] | got: [%v]", name, meta.Name)
	}

	piece_length := int64(262144)
	if meta.PieceLength != piece_length {
		t.Errorf("expected: [%v] | got: [%v]", piece_length, meta.PieceLength)
	}

	pieces := 235340
	if len(meta.Pieces) != pieces {
		t.Errorf("expected: [%v] | got: [%v]", pieces, len(meta.Pieces))
	}

	if len(meta.Files) != 2 {
		t.Fatalf("expected: [%v files] | got: [%v files]", 2, len(meta.Files))
	}

	path1 := "Fedora-Budgie-Live-44-1.7.x86_64.iso"
	path2 := "Fedora-Spins-44-1.7-x86_64-CHECKSUM"
	length1 := int64(3084500992)
	length2 := int64(2922)

	if meta.Files[0].Path != path1 {
		t.Fatalf("expected: [%v] | got: [%v]", path1, meta.Files[0].Path)
	}
	if meta.Files[1].Path != path2 {
		t.Fatalf("expected: [%v] | got: [%v]", path2, meta.Files[1].Path)
	}
	if meta.Files[0].Length != length1 {
		t.Fatalf("expected: [%v] | got: [%v]", length1, meta.Files[0].Length)
	}

	if meta.Files[1].Length != length2 {
		t.Fatalf("expected: [%v] | got: [%v]", length2, meta.Files[1].Length)
	}

	infohash := [20]byte{
		0x78, 0xbe, 0x65, 0x13,
		0xb4, 0x49, 0xf8, 0x18,
		0x82, 0xc8, 0x52, 0x07,
		0x5c, 0xf0, 0x11, 0x61,
		0xf3, 0xb7, 0xb2, 0xa3,
	}

	if meta.Infohash != infohash {
		t.Errorf("expected: [%x] | got: [%x]", infohash, meta.Infohash)
	}
}
