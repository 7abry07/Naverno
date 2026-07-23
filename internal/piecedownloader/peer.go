package piecedownloader

import "Naverno/internal/piece"

type Peer interface {
	Request(pi *piece.Piece, begin, length uint32)
}
