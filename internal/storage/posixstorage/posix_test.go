package posixstorage_test

import (
	"Naverno/internal/metadata"
	"Naverno/internal/piece"
	"Naverno/internal/storage/posixstorage"
	"bytes"
	"crypto/sha1"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestStorage(t *testing.T) {
	dir := t.TempDir()
	files := []metadata.File{
		{Length: 6, Path: "f1.bin"},
		{Length: 3, Path: "f2.bin"},
		{Length: 9, Path: "f3.bin"},
	}
	writeData := make([]byte, 6)
	copy(writeData, []byte("hellow"))
	pieceHash := [20]byte{}
	hasher := sha1.New()
	hasher.Write(writeData)
	copy(pieceHash[:], hasher.Sum(nil))

	p1 := piece.NewPiece(0, 6, 0, pieceHash)

	os.WriteFile(filepath.Join(dir, files[0].Path), make([]byte, 6), 0644)
	os.WriteFile(filepath.Join(dir, files[1].Path), make([]byte, 3), 0644)
	os.WriteFile(filepath.Join(dir, files[2].Path), make([]byte, 9), 0644)

	s := posixstorage.New(slog.New(slog.NewTextHandler(io.Discard, nil)), files, dir)
	err := s.Write(p1, 0, writeData[:3])
	if err != nil {
		t.Fatalf("unexpected error -> %v", err)
	}

	readData, err := s.Read(p1, 0, 3)
	if err != nil {
		t.Fatalf("unexpected error -> %v", err)
	}

	if !bytes.Equal(writeData[:3], readData) {
		t.Errorf("data read is not equal to data written, expected -> %v, got -> %v", string(writeData), string(readData))
	}
	err = s.Write(p1, 3, writeData[3:6])
	if err != nil {
		t.Fatalf("unexpected error -> %v", err)
	}

	ok, err := s.Hash(p1)
	if err != nil {
		t.Fatalf("unexpected error -> %v", err)
	}

	if !ok {
		t.Errorf("piece hash doesn't match")
	}
}
