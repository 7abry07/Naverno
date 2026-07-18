package util

import (
	"encoding/binary"
)

func BytesToUint64s(data []byte) []uint64 {
	padded := make([]byte, (len(data)+7)&^7)
	copy(padded, data)

	out := make([]uint64, len(padded)/8)
	for i := range out {
		out[i] = binary.BigEndian.Uint64(padded[i*8:])
	}

	return out
}
