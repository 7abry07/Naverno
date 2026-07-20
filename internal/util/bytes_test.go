package util_test

import (
	"Naverno/internal/util"
	"reflect"
	"testing"
)

func TestBitsetConversion(t *testing.T) {
	data := []byte{0b00001000, 0b00000001}
	bs, err := util.BytesToBitset(data, 12)
	if err != nil {
		t.Fatalf("unexpected error -> %v", err)
	}
	b := util.BitsetToBytes(bs)
	if !reflect.DeepEqual(b, data) {
		t.Errorf("the bytes are not equal, expected -> %v, got -> %v", data, b)
	}

	expected := []uint{4, 15}
	setBits := []uint{}
	for i := range bs.EachSet() {
		setBits = append(setBits, i)
	}

	if !reflect.DeepEqual(expected, setBits) {
		t.Errorf("iterating set bits didn't return set bits, set -> %v, got -> %v", expected, setBits)
	}
}

func TestByteConversion(t *testing.T) {
	data := []byte{0b11111111, 0b01111111}
	uint64s := util.BytesToUint64s(data)
	if !reflect.DeepEqual(data, util.Uint64sToBytes(uint64s, len(data))) {
		t.Errorf("the bytes are not equal, expected -> %v, got -> %v", data, util.Uint64sToBytes(uint64s, len(data)))
	}
}

func TestByteAlign(t *testing.T) {
	if val := util.Align(12, 8); val != 16 {
		t.Errorf("alignement is wrong, %v -> %v == %v, got -> %v", 12, 8, 16, val)
	}
}
