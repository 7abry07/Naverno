package bitfield

import (
	"Naverno/internal/util"
	"encoding/binary"
	"fmt"
	"iter"
	"math/bits"

	"github.com/bits-and-blooms/bitset"
)

type Bitfield struct {
	set bitset.BitSet
}

func New(length uint32) *Bitfield {
	return &Bitfield{*bitset.MustNew(uint(length))}
}

func From(data []byte, length uint32) (*Bitfield, error) {
	minimumBits := util.Align(uint64(length), 8)
	if len(data)*8 != int(minimumBits) {
		return nil, fmt.Errorf("invalid length")
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

	return &Bitfield{*bitset.FromWithLength(uint(length), words)}, nil
}

func (b *Bitfield) Bytes() []byte {
	minimumStorage := util.Align(uint64(b.set.Len()), 8) / 8

	out := make([]byte, minimumStorage)
	buf := []byte{}

	for _, w := range b.set.Words() {
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

func (b *Bitfield) SetBits() iter.Seq[uint] {
	return b.set.EachSet()
}

func (b *Bitfield) Count() uint32 {
	return uint32(b.set.Count())
}

func (b *Bitfield) Len() uint32 {
	return uint32(b.set.Len())
}

func (b *Bitfield) All() bool {
	return b.set.All()
}

func (b *Bitfield) Set(i uint32) *Bitfield {
	b.set.Set(uint(i))
	return b
}

func (b *Bitfield) Clear(i uint32) *Bitfield {
	b.set.Clear(uint(i))
	return b
}

func (b *Bitfield) Test(i uint32) bool {
	return b.set.Test(uint(i))
}
