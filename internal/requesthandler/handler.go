package requesthandler

import (
	"Naverno/internal/peerprotocol"
	"Naverno/internal/piece"
	"Naverno/internal/storage"
)

type RequestHandler struct {
	Requester [20]byte
	Piece     *piece.Piece
	Request   peerprotocol.Request
	storage   storage.Storage
	Data      []byte
	Err       error

	closeC chan struct{}
	doneC  chan struct{}
}

func New(storage storage.Storage, requester [20]byte, piece *piece.Piece, req peerprotocol.Request) *RequestHandler {
	return &RequestHandler{
		Requester: requester,
		Request:   req,
		Piece:     piece,
		storage:   storage,
		closeC:    make(chan struct{}),
		doneC:     make(chan struct{}),
	}
}

func (h *RequestHandler) Run(results chan *RequestHandler) {
	defer close(h.doneC)
	defer func() {
		select {
		case <-h.closeC:
		case results <- h:
		}
	}()

	data, err := h.storage.Read(h.Piece.Offset+uint64(h.Request.Begin), h.Request.Length)
	h.Err = err
	h.Data = data
}

func (h *RequestHandler) Close() {
	close(h.closeC)
	<-h.doneC
}
