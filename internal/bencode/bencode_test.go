package bencode

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

func TestSimpleInt(t *testing.T) {
	input := "i56e"
	val, err := Decode(input)
	if err != nil {
		t.Errorf("expected: [%v] | got: [%v]", 56, err)
	}
	intval, ok := val.Int()
	if !ok {
		t.Errorf("expected: [%v] | got: [%v]", 56, "not int")
	}
	if intval != 56 {
		t.Errorf("expected: [%v] | got: [%v]", 56, intval)
	}
}

func TestSimpleString(t *testing.T) {
	input := "5:hello"
	val, err := Decode(input)
	if err != nil {
		t.Errorf("expected: [%v] | got: [%v]", "'hello'", err)
	}
	strval, ok := val.Str()
	if !ok {
		t.Errorf("expected: [%v] | got: [%v]", "'hello'", "not string")
	}
	if strval != "hello" {
		t.Errorf("expected: [%v] | got: [%v]", "'hello'", strval)
	}
}

func TestSimpleList(t *testing.T) {
	input := "l5:helloi56ee"
	val, err := Decode(input)
	if err != nil {
		t.Errorf("expected: [%v] | got: [%v]", "list", err)
	}
	listval, ok := val.List()
	if !ok {
		t.Errorf("expected: [%v] | got: [%v]", "list", "not list")
	}

	stritem, ok := listval[0].Str()
	intitem, ok2 := listval[1].Int()

	if !ok {
		t.Errorf("expected item: %v->[%v] | got: [%v]", 0, "hello", "not string")
	}
	if !ok2 {
		t.Errorf("expected item: %v->[%v] | got: [%v]", 1, 56, "not int")
	}

	if stritem != "hello" {
		t.Errorf("expected item: %v->[%v] | got: [%v]", 0, "hello", stritem)
	}
	if intitem != 56 {
		t.Errorf("expected item: %v->[%v] | got: [%v]", 1, 56, intitem)
	}
}

func TestSimpleDict(t *testing.T) {
	input := "d5:helloi56ee"
	val, err := Decode(input)
	if err != nil {
		t.Errorf("expected: [%v] | got: [%v]", "dict", err)
	}
	dictval, ok := val.Dict()
	if !ok {
		t.Errorf("expected: [%v] | got: [%v]", "dict", "not dict")
	}

	value, ok := dictval.FindInt("hello")
	if !ok {
		t.Errorf("expected key-value: [%v-%v] | got: [%v]", "hello", 56, "key not found")
	}
	if value != 56 {
		t.Errorf("expected key-value: [%v-%v] | got: [%v-%v]", "hello", 56, "hello", value)
	}
}

func TestEmptyInput(t *testing.T) {
	input := ""
	_, err := Decode(input)
	if err != EmptyInputErr {
		t.Errorf("expected: [%v] | got: [%v]", EmptyInputErr, err)
	}
}

func TestMaxDepth(t *testing.T) {
	var input strings.Builder
	for i := range 200 {
		if i < 100 {
			input.WriteByte('l')
		} else {
			input.WriteByte('e')
		}
	}
	_, err := Decode(input.String())
	if err != MaximumNestingErr {
		t.Errorf("expected: [%v] | got: [%v]", MaximumNestingErr, err)
	}
}

func TestInvalidType(t *testing.T) {
	input := "t"
	_, err := Decode(input)
	if err != InvalidTypeErr {
		t.Errorf("expected: [%v] | got: [%v]", InvalidTypeErr, err)
	}
}

func TestTrailingInput(t *testing.T) {
	input := "5:hellotrailing"
	_, err := Decode(input)
	if err != TrailingInputErr {
		t.Errorf("expected: [%v] | got: [%v]", TrailingInputErr, err)
	}
}

func TestInvalidInteger(t *testing.T) {
	input := fmt.Sprintf("i%v0e", math.MaxInt64)
	_, err := Decode(input)
	if err != InvalidIntErr {
		t.Errorf("expected: [%v] | got: [%v]", InvalidIntErr, err)
	}

	input = fmt.Sprintf("i%v0e", math.MinInt64)
	_, err2 := Decode(input)
	if err2 != InvalidIntErr {
		t.Errorf("expected: [%v] | got: [%v]", InvalidIntErr, err2)
	}
}

func TestMissingIntTerm(t *testing.T) {
	input := "i65"
	_, err := Decode(input)
	if err != MissingIntTermErr {
		t.Errorf("expected: [%v] | got: [%v]", MissingIntTermErr, err)
	}
}

func TestInvalidStringLength(t *testing.T) {
	input := "5h3:hello"
	_, err := Decode(input)
	if err != InvalidStrLengthErr {
		t.Errorf("expected: %v | got: %v", InvalidStrLengthErr, err)
	}
}

func TestLengthMismatch(t *testing.T) {
	input := "5:hell"
	_, err := Decode(input)
	if err != LengthMismatchErr {
		t.Errorf("expected: [%v] | got: [%v]", LengthMismatchErr, err)
	}
}

func TestMissingColon(t *testing.T) {
	input := "5hello"
	_, err := Decode(input)
	if err != MissingColonErr {
		t.Errorf("expected: [%v] | got: [%v]", MissingColonErr, err)
	}
}

func TestMissingListTerm(t *testing.T) {
	input := "l5:hello"
	_, err := Decode(input)
	if err != MissingListTermErr {
		t.Errorf("expected: [%v] | got: [%v]", MissingListTermErr, err)
	}
}

func TestMissingDictTerm(t *testing.T) {
	input := "d5:helloi56e"
	_, err := Decode(input)
	if err != MissingDictTermErr {
		t.Errorf("expected: [%v] | got: [%v]", MissingDictTermErr, err)
	}
}

func TestNonStrKey(t *testing.T) {
	input := "di56e5:helloe"
	_, err := Decode(input)
	if err != NonStrKeyErr {
		t.Errorf("expected: [%v] | got: [%v]", NonStrKeyErr, err)
	}
}

func TestNonSortedKey(t *testing.T) {
	input := "d4:zetai10e5:alphai10ee"
	_, err := Decode(input)
	if err != NonSortedKeysErr {
		t.Errorf("expected: [%v] | got: [%v]", NonSortedKeysErr, err)
	}
}

func TestDuplicateKey(t *testing.T) {
	input := "d5:hello2:hi5:helloi43ee"
	_, err := Decode(input)
	if err != DuplicateKeyErr {
		t.Errorf("expected: [%v] | got: [%v]", DuplicateKeyErr, err)
	}
}

//
// ENCODER
//

func TestSimpleEncoding(t *testing.T) {
	input := "d4:helli43e5:hello2:hie"
	val, err := Decode(input)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	encoded := Encode(val)
	if encoded != input {
		t.Errorf("expected: [%v] | got: [%v]", input, encoded)
	}
}

