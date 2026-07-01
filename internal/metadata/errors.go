package metadata

import "errors"

var (
	RootNotDictErr     = errors.New("the root structure is not a dictionary")
	MissingAnnounceErr = errors.New("the mandatory field 'announce' or 'announce-list' is missing")
	MissingInfoErr     = errors.New("the mandatory field 'info' is missing")
	MissingNameErr     = errors.New("the mandatory field 'name' is missing")
	MissingPiecesErr   = errors.New("the mandatory field 'pieces' is missing")
	MissingPieceLenErr = errors.New("the mandatory field 'piece length' is missing")

	InvalidAnnounceErr        = errors.New("the 'announce' field is invalid")
	InvalidPiecesErr          = errors.New("the 'pieces' field is invalid")
	InvalidFilesErr           = errors.New("the 'files' field is invalid")
	BothLengthFilesPresentErr = errors.New("both 'files' and 'length' fields are present")
	BothLengthFilesMissingErr = errors.New("both 'files' and 'length' fields are missing")
)
