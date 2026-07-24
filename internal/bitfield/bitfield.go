package bitfield

import (
	"Naverno/internal/util"
	"encoding/binary"
	"fmt"
	"math/bits"

	"github.com/bits-and-blooms/bitset"
)

type Bitfield struct {
	*bitset.BitSet
}

func New(length uint32) Bitfield {
	return Bitfield{bitset.MustNew(uint(length))}
}

func From(data []byte, length uint32) (Bitfield, error) {
	spareBits := uint32(len(data)*8) - length
	for i := range spareBits {
		if data[len(data)-1]&1<<i != 0 {
			return Bitfield{}, fmt.Errorf("spare bits are set")
		}
	}

	minimumBits := util.Align(uint64(length), 8)
	if len(data)*8 != int(minimumBits) {
		return Bitfield{}, fmt.Errorf("invalid length")
	}
	buf := make([]byte, minimumBits/8)
	copy(buf, data)

	padded := make([]byte, util.Align(uint64(len(buf)), 64))
	copy(padded, buf)

	for i := range padded {
		padded[i] = bits.Reverse8(padded[i])
	}

	words := make([]uint64, len(padded)/8)
	for i := range words {
		words[i] = binary.LittleEndian.Uint64(padded[i*8:])
	}

	return Bitfield{bitset.FromWithLength(uint(length), words)}, nil
}

func (b *Bitfield) Bytes() []byte {
	minimumStorage := util.Align(uint64(b.Len()), 8) / 8

	out := make([]byte, minimumStorage)
	buf := []byte{}

	for _, w := range b.Words() {
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
