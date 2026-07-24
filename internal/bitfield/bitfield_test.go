package bitfield_test

import (
	"Naverno/internal/bitfield"
	"reflect"
	"testing"
)

func TestBitifled(t *testing.T) {
	data := []byte{0b00000001, 0b0001000}
	bs, err := bitfield.From(data, 12)
	if err != nil {
		t.Fatalf("unexpected error -> %v", err)
	}
	b := bs.Bytes()
	if !reflect.DeepEqual(b, data) {
		t.Errorf("the bytes are not equal, expected -> %v, got -> %v", data, b)
	}

	expected := []uint{7, 12}
	setBits := []uint{}
	for i := range bs.EachSet() {
		setBits = append(setBits, i)
	}

	if !reflect.DeepEqual(expected, setBits) {
		t.Errorf("iterating set bits didn't return set bits, set -> %v, got -> %v", expected, setBits)
	}
}

func TestSpareBits(t *testing.T) {
	data := []byte{0b00000001, 0b00000111}
	_, err := bitfield.From(data, 12)
	if err == nil {
		t.Errorf("should have returned error")
	}
}
