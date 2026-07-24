package filestorage

import (
	"Naverno/internal/metadata"
	"cmp"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
)

type File struct {
	metadata.File
	Offset uint64
}

type FileStorage struct {
	logger *slog.Logger
	files  []File
	path   string
}

func New(logger *slog.Logger, files []metadata.File, path string) *FileStorage {
	if logger == nil {
		panic("passed nil logger to file storage")
	}

	offs := []File{}
	off := uint64(0)
	for _, f := range files {
		offs = append(offs, File{File: f, Offset: off})
		off += uint64(f.Length)
	}

	for _, f := range offs {
		logger.Info("file", "file", f)
	}

	slices.SortFunc(offs, func(e1, e2 File) int { return cmp.Compare(e1.Offset, e2.Offset) })
	return &FileStorage{
		logger: logger,
		files:  offs,
		path:   path,
	}
}

func (s *FileStorage) Write(off uint64, data []byte) error {
	for _, f := range s.files {
		if len(data) == 0 {
			break
		}
		if off >= f.Offset {
			fileOff := off - f.Offset
			writeLen := min(len(data), int(uint64(f.Length)-fileOff))
			if writeLen == 0 {
				continue
			}
			err := os.MkdirAll(s.path, 0755)
			if err != nil {
				return err
			}
			handle, err := os.OpenFile(filepath.Join(s.path, f.Path), os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			_, err = handle.WriteAt(data[:writeLen], int64(fileOff))
			if err != nil {
				return err
			}
			s.logger.Debug("file storage -> written data", "File", f.Path, "GlobalOff", f.Offset, "LocalOff", fileOff, "DataOffs", off, "Length", writeLen)
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
			readLen := min(length, uint32(uint64(f.Length)-fileOff))
			if readLen == 0 {
				continue
			}
			buf := make([]byte, readLen)
			err := os.MkdirAll(s.path, 0755)
			if err != nil {
				return []byte{}, err
			}
			handle, err := os.OpenFile(filepath.Join(s.path, f.Path), os.O_CREATE|os.O_RDONLY, 0644)
			if err != nil {
				return []byte{}, err
			}
			_, err = handle.ReadAt(buf, int64(fileOff))
			if err != nil {
				return []byte{}, err
			}
			s.logger.Debug("file storage -> read data", "File", f.Path, "GlobalOff", f.Offset, "LocalOff", fileOff, "DataOffs", off, "Length", readLen)
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
