package piecewriter

import (
	"Naverno/internal/piece"
	"Naverno/internal/storage"
)

type PieceWriter struct {
	Piece   *piece.Piece
	Begin   uint32
	data    []byte
	storage storage.Storage
	Err     error

	closeC chan struct{}
	doneC  chan struct{}
}

func New(p *piece.Piece, begin uint32, s storage.Storage, data []byte) *PieceWriter {
	return &PieceWriter{
		Piece:   p,
		Begin:   begin,
		storage: s,
		data:    data,

		closeC: make(chan struct{}),
		doneC:  make(chan struct{}),
	}
}

func (w *PieceWriter) Run(result chan<- *PieceWriter) {
	defer func() {
		select {
		case <-w.closeC:
		case result <- w:
		}
	}()

	err := w.storage.Write(w.Piece.Offset+uint64(w.Begin), w.data)
	w.Err = err
	result <- w
}

func (w *PieceWriter) Close() {
	close(w.closeC)
	<-w.doneC
}
