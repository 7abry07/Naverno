package util

import (
	"encoding/binary"
	"fmt"

	"github.com/bits-and-blooms/bitset"
)

func BytesToBitset(data []byte, bits uint) (*bitset.BitSet, error) {
	minimumBits := Align(uint64(bits), 8)
	if len(data)*8 != int(minimumBits) {
		return nil, fmt.Errorf("invalid length")
	}
	buf := make([]byte, minimumBits/8)
	copy(buf, data)

	return bitset.FromWithLength(bits, BytesToUint64s(buf)), nil
}

func BitsetToBytes(bs *bitset.BitSet) []byte {
	minimumStorage := Align(uint64(bs.Len()), 8) / 8
	b := Uint64sToBytes(bs.Words(), int(minimumStorage))
	return b
}

func Align(n, alignment uint64) uint64 {
	return (n + alignment - 1) / alignment * alignment
}

func BytesToUint64s(data []byte) []uint64 {
	padded := make([]byte, Align(uint64(len(data)), 64))
	copy(padded, data)

	out := make([]uint64, len(padded)/8)
	for i := range out {
		out[i] = binary.LittleEndian.Uint64(padded[i*8:])
	}

	return out
}

func Uint64sToBytes(data []uint64, datalen int) []byte {
	out := make([]byte, datalen)
	buf := []byte{}

	for _, w := range data {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, w)
		buf = append(buf, b...)
	}

	if len(buf) < datalen {
		panic(fmt.Errorf("slice is too short, slice length -> %v | length -> %v", len(buf), datalen))
	}

	copy(out, buf[:datalen])
	return out
}
