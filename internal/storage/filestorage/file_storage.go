package filestorage

import (
	"Naverno/internal/metadata"
	"cmp"
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

type File struct {
	metadata.File
	Offset uint64
}

type FileStorage struct {
	files []File
	path  string
}

func New(files []metadata.File, path string) *FileStorage {
	offs := []File{}
	off := uint64(0)
	for _, f := range files {
		offs = append(offs, File{File: f, Offset: off})
		off += uint64(f.Length)
	}

	slices.SortFunc(offs, func(e1, e2 File) int { return cmp.Compare(e1.Offset, e2.Offset) })
	return &FileStorage{
		files: offs,
		path:  path,
	}
}

func (s *FileStorage) Write(off uint64, data []byte) error {
	for _, f := range s.files {
		if len(data) == 0 {
			break
		}
		if off >= f.Offset {
			fileOff := off - f.Offset
			writeLen := min(len(data), int(f.Length))
			handle, err := os.OpenFile(filepath.Join(s.path, f.Path), os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			_, err = handle.WriteAt(data[:writeLen], int64(fileOff))
			if err != nil {
				return err
			}
			data = data[writeLen:]
			off += uint64(writeLen)
			handle.Close()
		}
	}
	if len(data) != 0 {
		return fmt.Errorf("couldn't write all data")
	}
	return nil
}

func (s *FileStorage) Read(off uint64, length uint32) ([]byte, error) {
	readData := []byte{}

	for _, f := range s.files {
		if length == 0 {
			break
		}
		if off >= f.Offset {
			fileOff := off - f.Offset
			readLen := min(length, uint32(f.Length))
			buf := make([]byte, readLen)
			handle, err := os.OpenFile(filepath.Join(s.path, f.Path), os.O_CREATE|os.O_RDONLY, 0644)
			if err != nil {
				return []byte{}, err
			}
			_, err = handle.ReadAt(buf, int64(fileOff))
			if err != nil {
				return []byte{}, err
			}
			readData = append(readData, buf...)
			length -= readLen
			off += uint64(readLen)
		}
	}

	if length != 0 {
		return []byte{}, fmt.Errorf("couldn't read all data")
	}
	return readData, nil
}
