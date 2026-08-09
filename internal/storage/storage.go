package storage

import "Naverno/internal/piece"

type Storage interface {
	Write(*piece.Piece, uint32, []byte) error
	Read(*piece.Piece, uint32, uint32) ([]byte, error)
	Hash(*piece.Piece) (bool, error)

	AsyncWrite(chan WriteResult, *piece.Piece, uint32, []byte)
	AsyncRead(chan ReadResult, *piece.Piece, uint32, uint32)
	AsyncHash(chan HashResult, *piece.Piece)

	StopPendingJobs()
}
