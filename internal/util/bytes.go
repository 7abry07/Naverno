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

func Uint64sToBytes(data []uint64, bits int) []byte {
	out := make([]byte, bits/8)
	buf := make([]byte, len(data)/8)

	for _, w := range data {
		b := make([]byte, 8)
		binary.BigEndian.PutUint64(b, w)
		buf = append(buf, b...)
	}

	if len(buf)*8 < bits {
		panic("slice is too short")
	}

	if bits%8 != 0 {
		panic("bit count isn't a multiple of 8")
	}

	copy(out, buf[:bits/8])
	return out
}
