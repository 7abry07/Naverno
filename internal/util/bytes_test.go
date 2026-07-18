package util_test

import (
	"Naverno/internal/util"
	"reflect"
	"testing"
)

func TestByteConversion(t *testing.T) {
	data := []byte{0b11111111, 0b01111111}
	uint64s := util.BytesToUint64s(data)
	if !reflect.DeepEqual(data, util.Uint64sToBytes(uint64s, len(data)*8)) {
		t.Error("the bytes are not equal")
	}
}
