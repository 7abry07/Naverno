package bencode

import "errors"

var (
	EmptyInputErr     = errors.New("the input is empty")
	InvalidTypeErr    = errors.New("invalid type specifier encountered")
	MaximumNestingErr = errors.New("maximum nesting limit excedeed")
	TrailingInputErr  = errors.New("trailing input not allowed")

	InvalidIntErr     = errors.New("invalid integer encountered")
	MissingIntTermErr = errors.New("integer terminator not found")

	InvalidStrLengthErr = errors.New("the string length is invalid")
	LengthMismatchErr   = errors.New("the length of the string doesn't match the payload")
	MissingColonErr     = errors.New("the colon between length and payload is missing")

	MissingListTermErr = errors.New("list terminator not found")

	MissingDictTermErr = errors.New("dictionary terminator not found")
	NonStrKeyErr       = errors.New("key to dictionary value not a string")
	DuplicateKeyErr    = errors.New("key to dictionary value already exists")
	NonSortedKeysErr   = errors.New("keys in the dictionary aren't sorted lexicographically")
)
