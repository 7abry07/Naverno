package bencode_test

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"Naverno/internal/bencode"
)

func TestSimpleInt(t *testing.T) {
	input := "i56e"
	val, err := bencode.Decode(input)
	if err != nil {
		t.Errorf("expected: [%v] | got: [%v]", 56, err)
	}
	intval, ok := val.(int64)
	if !ok {
		t.Errorf("expected: [%v] | got: [%v]", 56, "not int")
	}
	if intval != 56 {
		t.Errorf("expected: [%v] | got: [%v]", 56, intval)
	}
}

func TestSimpleString(t *testing.T) {
	input := "5:hello"
	val, err := bencode.Decode(input)
	if err != nil {
		t.Errorf("expected: [%v] | got: [%v]", "'hello'", err)
	}
	strval, ok := val.(string)
	if !ok {
		t.Errorf("expected: [%v] | got: [%v]", "'hello'", "not string")
	}
	if strval != "hello" {
		t.Errorf("expected: [%v] | got: [%v]", "'hello'", strval)
	}
}

func TestSimpleList(t *testing.T) {
	input := "l5:helloi56ee"
	val, err := bencode.Decode(input)
	if err != nil {
		t.Errorf("expected: [%v] | got: [%v]", "list", err)
	}
	listval, ok := val.([]any)
	if !ok {
		t.Errorf("expected: [%v] | got: [%v]", "list", "not list")
	}

	stritem, ok := listval[0].(string)
	intitem, ok2 := listval[1].(int64)

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
	val, err := bencode.Decode(input)
	if err != nil {
		t.Errorf("expected: [%v] | got: [%v]", "dict", err)
	}
	dictval, ok := val.(map[string]any)
	if !ok {
		t.Errorf("expected: [%v] | got: [%v]", "dict", "not dict")
	}

	value, ok := dictval["hello"]
	if !ok {
		t.Errorf("expected key-value: [%v-%v] | got: [%v]", "hello", 56, "key not found")
	}

	intvalue, ok := value.(int64)
	if !ok {
		t.Errorf("expected key-value: [%v-%v] | got: [%v]", "hello", 56, "value not integer")
	}

	if intvalue != 56 {
		t.Errorf("expected key-value: [%v-%v] | got: [%v-%v]", "hello", 56, "hello", value)
	}
}

func TestEmptyInput(t *testing.T) {
	input := ""
	_, err := bencode.Decode(input)
	if err != bencode.EmptyInputErr {
		t.Errorf("expected: [%v] | got: [%v]", bencode.EmptyInputErr, err)
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
	_, err := bencode.Decode(input.String())
	if err != bencode.MaximumNestingErr {
		t.Errorf("expected: [%v] | got: [%v]", bencode.MaximumNestingErr, err)
	}
}

func TestInvalidType(t *testing.T) {
	input := "t"
	_, err := bencode.Decode(input)
	if err != bencode.InvalidTypeErr {
		t.Errorf("expected: [%v] | got: [%v]", bencode.InvalidTypeErr, err)
	}
}

func TestTrailingInput(t *testing.T) {
	input := "5:hellotrailing"
	_, err := bencode.Decode(input)
	if err != bencode.TrailingInputErr {
		t.Errorf("expected: [%v] | got: [%v]", bencode.TrailingInputErr, err)
	}
}

func TestInvalidInteger(t *testing.T) {
	input := fmt.Sprintf("i%v0e", math.MaxInt64)
	_, err := bencode.Decode(input)
	if err != bencode.InvalidIntErr {
		t.Errorf("expected: [%v] | got: [%v]", bencode.InvalidIntErr, err)
	}

	input = fmt.Sprintf("i%v0e", math.MinInt64)
	_, err2 := bencode.Decode(input)
	if err2 != bencode.InvalidIntErr {
		t.Errorf("expected: [%v] | got: [%v]", bencode.InvalidIntErr, err2)
	}
}

func TestMissingIntTerm(t *testing.T) {
	input := "i65"
	_, err := bencode.Decode(input)
	if err != bencode.MissingIntTermErr {
		t.Errorf("expected: [%v] | got: [%v]", bencode.MissingIntTermErr, err)
	}
}

func TestInvalidStringLength(t *testing.T) {
	input := "5h3:hello"
	_, err := bencode.Decode(input)
	if err != bencode.InvalidStrLengthErr {
		t.Errorf("expected: %v | got: %v", bencode.InvalidStrLengthErr, err)
	}
}

func TestLengthMismatch(t *testing.T) {
	input := "5:hell"
	_, err := bencode.Decode(input)
	if err != bencode.LengthMismatchErr {
		t.Errorf("expected: [%v] | got: [%v]", bencode.LengthMismatchErr, err)
	}
}

func TestMissingColon(t *testing.T) {
	input := "5hello"
	_, err := bencode.Decode(input)
	if err != bencode.MissingColonErr {
		t.Errorf("expected: [%v] | got: [%v]", bencode.MissingColonErr, err)
	}
}

func TestMissingListTerm(t *testing.T) {
	input := "l5:hello"
	_, err := bencode.Decode(input)
	if err != bencode.MissingListTermErr {
		t.Errorf("expected: [%v] | got: [%v]", bencode.MissingListTermErr, err)
	}
}

func TestMissingDictTerm(t *testing.T) {
	input := "d5:helloi56e"
	_, err := bencode.Decode(input)
	if err != bencode.MissingDictTermErr {
		t.Errorf("expected: [%v] | got: [%v]", bencode.MissingDictTermErr, err)
	}
}

func TestNonStrKey(t *testing.T) {
	input := "di56e5:helloe"
	_, err := bencode.Decode(input)
	if err != bencode.NonStrKeyErr {
		t.Errorf("expected: [%v] | got: [%v]", bencode.NonStrKeyErr, err)
	}
}

func TestNonSortedKey(t *testing.T) {
	input := "d4:zetai10e5:alphai10ee"
	_, err := bencode.Decode(input)
	if err != bencode.NonSortedKeysErr {
		t.Errorf("expected: [%v] | got: [%v]", bencode.NonSortedKeysErr, err)
	}
}

func TestDuplicateKey(t *testing.T) {
	input := "d5:hello2:hi5:helloi43ee"
	_, err := bencode.Decode(input)
	if err != bencode.DuplicateKeyErr {
		t.Errorf("expected: [%v] | got: [%v]", bencode.DuplicateKeyErr, err)
	}
}

//
// ENCODER
//

func TestSimpleEncoding(t *testing.T) {
	input := "d4:helli43e5:hello2:hie"
	val, err := bencode.Decode(input)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	encoded := bencode.Encode(val)
	if encoded != input {
		t.Errorf("expected: [%v] | got: [%v]", input, encoded)
	}
}
