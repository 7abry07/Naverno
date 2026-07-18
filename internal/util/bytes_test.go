package util_test

import (
	"Naverno/internal/util"
	"reflect"
	"testing"
)

func TestBitsetConversion(t *testing.T) {
	data := []byte{0b11111111, 0b01111111}
	bs, err := util.BytesToBitset(data, 12)
	if err != nil {
		t.Fatalf("unexpected error -> %v", err)
	}
	b := util.BitsetToBytes(bs)
	if !reflect.DeepEqual(b, data) {
		t.Errorf("the bytes are not equal, expected -> %v, got -> %v", data, b)
	}
}

func TestByteConversion(t *testing.T) {
	data := []byte{0b11111111, 0b01111111}
	uint64s := util.BytesToUint64s(data)
	if !reflect.DeepEqual(data, util.Uint64sToBytes(uint64s, len(data))) {
		t.Errorf("the bytes are not equal, expected -> %v, got -> %v", data, util.Uint64sToBytes(uint64s, len(data)))
	}
}
