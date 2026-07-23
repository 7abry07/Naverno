package picker

import (
	"Naverno/internal/piece"
)

type PieceState uint8

const (
	PIECE_FREE PieceState = iota
	PIECE_DOWNLOADING
	PIECE_COMPLETED
)

type Piece struct {
	*piece.Piece
	State PieceState
}
