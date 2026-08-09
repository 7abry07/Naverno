package posixstorage

import (
	"Naverno/internal/metadata"
	"Naverno/internal/piece"
	"Naverno/internal/storage"
	"bytes"
	"cmp"
	"crypto/sha1"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"sync"
)

type File struct {
	metadata.File
	Offset uint64
}

type PosixStorage struct {
	logger *slog.Logger
	files  []File
	path   string

	wg     sync.WaitGroup
	closeC chan struct{}
}

func New(logger *slog.Logger, files []metadata.File, path string) *PosixStorage {
	if logger == nil {
		panic("passed nil logger to file storage")
	}

	offs := []File{}
	off := uint64(0)
	for _, f := range files {
		offs = append(offs, File{File: f, Offset: off})
		off += uint64(f.Length)
	}

	slices.SortFunc(offs, func(e1, e2 File) int { return cmp.Compare(e1.Offset, e2.Offset) })
	return &PosixStorage{
		logger: logger,
		files:  offs,
		path:   path,
		wg:     sync.WaitGroup{},
		closeC: make(chan struct{}),
	}
}

func (s *PosixStorage) Write(p *piece.Piece, begin uint32, data []byte) error {
	off := uint64(p.Offset) + uint64(begin)

	for _, f := range s.files {
		fileEnd := f.Offset + uint64(f.Length)
		if off >= f.Offset && off < fileEnd {
			fileOff := off - f.Offset
			writeLen := min(len(data), int(uint64(f.Length)-fileOff))
			if writeLen == 0 {
				break
			}
			path := filepath.Join(s.path, f.Path)
			err := os.MkdirAll(filepath.Dir(path), 0755)
			if err != nil {
				return err
			}
			handle, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
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

func (s *PosixStorage) Read(p *piece.Piece, begin, length uint32) ([]byte, error) {
	off := uint64(p.Offset) + uint64(begin)
	length_ := uint64(length)
	readData := []byte{}

	for _, f := range s.files {
		fileEnd := f.Offset + uint64(f.Length)
		if off >= f.Offset && off < fileEnd {
			fileOff := off - f.Offset
			readLen := min(length_, uint64(f.Length)-fileOff)
			if readLen == 0 {
				break
			}
			buf := make([]byte, readLen)

			path := filepath.Join(s.path, f.Path)
			err := os.MkdirAll(filepath.Dir(path), 0755)
			if err != nil {
				return []byte{}, err
			}
			handle, err := os.OpenFile(path, os.O_CREATE|os.O_RDONLY, 0644)
			if err != nil {
				return []byte{}, err
			}
			_, err = handle.ReadAt(buf, int64(fileOff))
			if err != nil {
				return []byte{}, err
			}
			s.logger.Debug("file storage -> read data", "File", f.Path, "GlobalOff", f.Offset, "LocalOff", fileOff, "DataOffs", off, "Length", readLen)
			readData = append(readData, buf...)
			length_ -= readLen
			off += uint64(readLen)
		}
	}

	if length_ != 0 {
		return []byte{}, fmt.Errorf("couldn't read all data")
	}
	return readData, nil
}

func (s *PosixStorage) Hash(p *piece.Piece) (bool, error) {
	off := uint64(p.Offset)
	length := uint64(p.Size)
	readData := []byte{}

	for _, f := range s.files {
		fileEnd := f.Offset + uint64(f.Length)
		if off >= f.Offset && off < fileEnd {
			fileOff := off - f.Offset
			readLen := min(length, uint64(f.Length)-fileOff)
			if readLen == 0 {
				break
			}
			buf := make([]byte, readLen)

			path := filepath.Join(s.path, f.Path)
			err := os.MkdirAll(filepath.Dir(path), 0755)
			if err != nil {
				return false, err
			}
			handle, err := os.OpenFile(path, os.O_CREATE|os.O_RDONLY, 0644)
			if err != nil {
				return false, err
			}
			_, err = handle.ReadAt(buf, int64(fileOff))
			if err != nil {
				return false, err
			}
			s.logger.Debug("file storage -> read data", "File", f.Path, "GlobalOff", f.Offset, "LocalOff", fileOff, "DataOffs", off, "Length", readLen)
			readData = append(readData, buf...)
			length -= readLen
			off += uint64(readLen)
		}
	}

	if length != 0 {
		return false, fmt.Errorf("couldn't read all data")
	}

	expected := sha1.New()
	expected.Write(readData)
	return bytes.Equal(expected.Sum(nil), p.Hash[:]), nil
}

func (s *PosixStorage) AsyncWrite(resC chan storage.WriteResult, p *piece.Piece, begin uint32, data []byte) {
	s.wg.Go(
		func() {
			res := storage.WriteResult{}
			defer func() {
				select {
				case resC <- res:
				case <-s.closeC:
				}
			}()

			err := s.Write(p, begin, data)
			res.Err = err
			res.Piece = p
			res.Begin = begin
			res.DataWritten = uint64(len(data))
		})
}

func (s *PosixStorage) AsyncRead(resC chan storage.ReadResult, p *piece.Piece, begin, length uint32) {
	s.wg.Go(
		func() {
			res := storage.ReadResult{}
			defer func() {
				select {
				case resC <- res:
				case <-s.closeC:
				}
			}()

			data, err := s.Read(p, begin, length)
			res.Err = err
			res.Piece = p
			res.Begin = begin
			res.Data = data
		})
}

func (s *PosixStorage) AsyncHash(resC chan storage.HashResult, p *piece.Piece) {
	s.wg.Go(
		func() {
			res := storage.HashResult{}
			defer func() {
				select {
				case resC <- res:
				case <-s.closeC:
				}
			}()

			ok, err := s.Hash(p)
			res.Err = err
			res.Piece = p
			res.Ok = ok
		})
}

func (s *PosixStorage) StopPendingJobs() {
	close(s.closeC)
	s.wg.Wait()
}
