package discardstorage

import (
	"Naverno/internal/piece"
	"Naverno/internal/storage"
	"sync"
)

type DiscardStorage struct {
	wg     sync.WaitGroup
	closeC chan struct{}
}

func New() *DiscardStorage {
	return &DiscardStorage{}
}

func (s *DiscardStorage) Write(p *piece.Piece, begin uint32, data []byte) error {
	return nil
}

func (s *DiscardStorage) Read(p *piece.Piece, begin, length uint32) ([]byte, error) {
	return make([]byte, length), nil
}

func (s *DiscardStorage) Hash(p *piece.Piece) (bool, error) {
	return true, nil
}

func (s *DiscardStorage) AsyncWrite(resC chan storage.WriteResult, p *piece.Piece, begin uint32, data []byte) {
	s.wg.Add(1)
	defer s.wg.Done()
	select {
	case <-s.closeC:
	case resC <- storage.WriteResult{
		Err:         nil,
		Piece:       p,
		Begin:       begin,
		DataWritten: uint64(len(data)),
	}:
	}

}

func (s *DiscardStorage) AsyncRead(resC chan storage.ReadResult, p *piece.Piece, begin, length uint32) {
	s.wg.Add(1)
	defer s.wg.Done()
	select {
	case <-s.closeC:
	case resC <- storage.ReadResult{
		Err:   nil,
		Piece: p,
		Begin: begin,
		Data:  make([]byte, length),
	}:
	}

}

func (s *DiscardStorage) AsyncHash(resC chan storage.HashResult, p *piece.Piece) {
	s.wg.Add(1)
	defer s.wg.Done()
	select {
	case <-s.closeC:
	case resC <- storage.HashResult{
		Err:   nil,
		Piece: p,
		Ok:    true,
	}:
	}
}

func (s *DiscardStorage) StopPendingJobs() {
	close(s.closeC)
	s.wg.Wait()
}
