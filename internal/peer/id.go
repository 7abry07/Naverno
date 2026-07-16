package peer

import (
	"encoding/binary"
	"fmt"
	"math/rand/v2"
)

func GenerateRandomID() [20]byte {
	id := make([]byte, 12)
	seed := [32]byte{}

	binary.PutUvarint(seed[:], uint64(rand.Uint32()))
	gen := rand.NewChaCha8(seed)
	_, err := gen.Read(id)
	if err != nil {
		panic(fmt.Errorf("error in generating client id -> %v", err))
	}

	full := []byte{}
	full = append(full, "-NV0001-"...)
	full = append(full, id...)
	result := [20]byte{}
	copy(result[:], full)

	return result
}
