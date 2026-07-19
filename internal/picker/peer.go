package picker

import "github.com/bits-and-blooms/bitset"

type Peer interface {
	GetPieces() *bitset.BitSet
}
