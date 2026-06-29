package bencode

import (
	"strconv"
	"strings"
)

const MaxDepth = 100

func Decode(input string) (BNode, error) {
	depth := 0
	value, err := decode(&input, &depth)
	if err != nil {
		return BNode{}, err
	}
	if len(input) != 0 {
		return BNode{}, TrailingInputErr
	}
	return value, nil
}

func decode(input *string, depth *int) (BNode, error) {
	*depth++

	if *depth == MaxDepth {
		return BNode{}, MaximumNestingErr
	}
	if len(*input) == 0 {
		return BNode{}, EmptyInputErr
	}

	c := (*input)[0]
	switch c {
	case '0':
		fallthrough
	case '1':
		fallthrough
	case '2':
		fallthrough
	case '3':
		fallthrough
	case '4':
		fallthrough
	case '5':
		fallthrough
	case '6':
		fallthrough
	case '7':
		fallthrough
	case '8':
		fallthrough
	case '9':
		{
			val, err := decodeStr(input)
			*depth--
			if err != nil {
				return BNode{}, err
			}
			return NewStr(val), nil
		}
	case 'i':
		{
			val, err := decodeInt(input)
			*depth--
			if err != nil {
				return BNode{}, err
			}
			return NewInt(val), nil
		}
	case 'l':
		{
			val, err := decodeList(input, depth)
			*depth--
			if err != nil {
				return BNode{}, err
			}
			return NewList(val), nil
		}
	case 'd':
		{
			val, err := decodeDict(input, depth)
			*depth--
			if err != nil {
				return BNode{}, err
			}
			return NewDict(val), nil
		}
	}
	return BNode{}, InvalidTypeErr
}

func decodeStr(input *string) (BStr, error) {
	lenEnd := strings.IndexByte(*input, ':')
	if lenEnd == -1 {
		return "", MissingColonErr
	}

	lenStr := (*input)[0:lenEnd]
	lenInt, err := strconv.ParseInt(lenStr, 10, 64)
	if err != nil {
		return "", InvalidStrLengthErr
	}
	*input = (*input)[lenEnd+1:]
	if int64(len(*input)) < lenInt {
		return "", LengthMismatchErr
	}

	payload := (*input)[0:lenInt]
	*input = (*input)[lenInt:]
	return BStr(payload), nil
}

func decodeInt(input *string) (BInt, error) {
	*input = (*input)[1:]
	intEnd := strings.IndexByte(*input, 'e')
	if intEnd == -1 {
		return 0, MissingIntTermErr
	}

	strVal := (*input)[0:intEnd]

	if len(strVal) > 1 && strVal[0] == '0' ||
		len(strVal) > 1 && strVal[0:2] == "-0" {
		return 0, InvalidIntErr
	}

	val, err := strconv.ParseInt(strVal, 10, 64)
	if err != nil {
		return 0, InvalidIntErr
	}

	*input = (*input)[intEnd+1:]
	return BInt(val), nil
}

func decodeList(input *string, depth *int) (BList, error) {
	*input = (*input)[1:]
	var list BList
	for {
		if len(*input) == 0 {
			return BList{}, MissingListTermErr
		}
		if (*input)[0] == 'e' {
			*input = (*input)[1:]
			return list, nil
		}

		val, err := decode(input, depth)
		if err != nil {
			return BList{}, err
		}

		list = append(list, val)
	}
}

func decodeDict(input *string, depth *int) (BDict, error) {
	*input = (*input)[1:]
	node := NewEmptyDict()
	dict, _ := node.Dict()

	previousKey := ""
	first := true

	for {
		if len(*input) == 0 {
			return BDict{}, MissingDictTermErr
		}
		if (*input)[0] == 'e' {
			*input = (*input)[1:]
			return dict, nil
		}

		keyNode, err := decode(input, depth)
		if err != nil {
			return BDict{}, err
		}

		key, ok := keyNode.Str()
		if !ok {
			return BDict{}, NonStrKeyErr
		}
		if key < BStr(previousKey) && !first {
			return BDict{}, NonSortedKeysErr
		}

		val, err := decode(input, depth)
		if err != nil {
			return BDict{}, err
		}

		_, exists := dict[string(key)]
		if exists {
			return BDict{}, DuplicateKeyErr
		}
		dict[string(key)] = val
		previousKey = string(key)
		first = false
	}
}
