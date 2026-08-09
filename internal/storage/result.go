package storage

import "Naverno/internal/piece"

type WriteResult struct {
	Piece       *piece.Piece
	Begin       uint32
	DataWritten uint64
	Err         error
}

type ReadResult struct {
	Piece *piece.Piece
	Begin uint32
	Data  []byte
	Err   error
}

type HashResult struct {
	Piece *piece.Piece
	Ok    bool
	Err   error
}
