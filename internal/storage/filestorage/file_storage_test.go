package filestorage_test

import (
	"Naverno/internal/metadata"
	"Naverno/internal/storage/filestorage"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestWrite(t *testing.T) {
	dir := t.TempDir()
	files := []metadata.File{
		{Length: 5, Path: "f1.bin"},
		{Length: 5, Path: "f2.bin"},
		{Length: 5, Path: "f3.bin"},
	}

	os.WriteFile(filepath.Join(dir, files[0].Path), make([]byte, 5), 0644)
	os.WriteFile(filepath.Join(dir, files[1].Path), make([]byte, 5), 0644)
	os.WriteFile(filepath.Join(dir, files[2].Path), make([]byte, 5), 0644)

	s := filestorage.New(files, dir)

	writeData := make([]byte, 7)
	copy(writeData, []byte("melodye"))
	err := s.Write(0, writeData)
	if err != nil {
		t.Fatalf("unexpected error -> %v", err)
	}

	readData, err := s.Read(0, 7)
	if err != nil {
		t.Fatalf("unexpected error -> %v", err)
	}

	if !bytes.Equal(writeData, readData) {
		t.Errorf("data read is not equal to data written, expected -> %v, got -> %v", string(writeData), string(readData))
	}
}
