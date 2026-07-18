package util_test

import (
	"Naverno/internal/util"
	"reflect"
	"testing"
)

func TestSliceRemove(t *testing.T) {
	slice := []string{"a", "b", "c", "d"}
	slice = util.Remove(slice, "c", func(e1, e2 string) bool { return e1 == e2 })
	expectedSlice := []string{"a", "b", "d"}
	if !reflect.DeepEqual(expectedSlice, slice) {
		t.Errorf("expected slice -> %v | got -> %v", expectedSlice, slice)
	}
}
