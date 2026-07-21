package util

import (
	"encoding/binary"
	"fmt"
	"math/bits"

	"github.com/bits-and-blooms/bitset"
)

func BytesToBitset(data []byte, datalen uint) (*bitset.BitSet, error) {
	minimumBits := Align(uint64(datalen), 8)
	if len(data)*8 != int(minimumBits) {
		return nil, fmt.Errorf("invalid length")
	}
	buf := make([]byte, minimumBits/8)
	copy(buf, data)

	padded := make([]byte, Align(uint64(len(buf)), 64))
	copy(padded, buf)

	for i := range padded {
		padded[i] = bits.Reverse8(padded[i])
	}

	words := make([]uint64, len(padded)/8)
	for i := range words {
		words[i] = binary.LittleEndian.Uint64(padded[i*8:])
	}

	return bitset.FromWithLength(datalen, words), nil
}

func BitsetToBytes(bs *bitset.BitSet) []byte {
	minimumStorage := Align(uint64(bs.Len()), 8) / 8

	out := make([]byte, minimumStorage)
	buf := []byte{}

	for _, w := range bs.Words() {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, w)
		buf = append(buf, b...)
	}

	for i := range buf {
		buf[i] = bits.Reverse8(buf[i])
	}

	if uint64(len(buf)) < minimumStorage {
		panic(fmt.Errorf("slice is too short, slice length -> %v | minimum length -> %v", len(buf), minimumStorage))
	}

	copy(out, buf[:minimumStorage])
	return out
}

func Align(n, alignment uint64) uint64 {
	return (n + alignment - 1) / alignment * alignment
}
