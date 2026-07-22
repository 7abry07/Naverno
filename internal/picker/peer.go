package picker

import (
	"Naverno/internal/bitfield"
)

type Peer interface {
	GetPieces() *bitfield.Bitfield
}
