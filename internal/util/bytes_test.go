package util_test

import (
	"Naverno/internal/util"
	"testing"
)

func TestAlign(t *testing.T) {
	if val := util.Align(12, 8); val != 16 {
		t.Errorf("alignement is wrong, %v -> %v == %v, got -> %v", 12, 8, 16, val)
	}
}
