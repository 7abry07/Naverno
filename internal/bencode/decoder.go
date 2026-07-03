package bencode

import (
	"strconv"
	"strings"
)

const MaxDepth = 100

func Decode(input string) (any, error) {
	depth := 0
	value, err := decode(&input, &depth)
	if err != nil {
		return nil, err
	}
	if len(input) != 0 {
		return nil, TrailingInputErr
	}
	return value, nil
}

func decode(input *string, depth *int) (any, error) {
	*depth++

	if *depth == MaxDepth {
		return nil, MaximumNestingErr
	}
	if len(*input) == 0 {
		return nil, EmptyInputErr
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
				return nil, err
			}
			return val, nil
		}
	case 'i':
		{
			val, err := decodeInt(input)
			*depth--
			if err != nil {
				return nil, err
			}
			return val, nil
		}
	case 'l':
		{
			val, err := decodeList(input, depth)
			*depth--
			if err != nil {
				return nil, err
			}
			return val, nil
		}
	case 'd':
		{
			val, err := decodeDict(input, depth)
			*depth--
			if err != nil {
				return nil, err
			}
			return val, nil
		}
	default:
		return nil, InvalidTypeErr
	}
}

func decodeStr(input *string) (string, error) {
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
	return payload, nil
}

func decodeInt(input *string) (int64, error) {
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
	return val, nil
}

func decodeList(input *string, depth *int) ([]any, error) {
	*input = (*input)[1:]
	var list []any
	for {
		if len(*input) == 0 {
			return nil, MissingListTermErr
		}
		if (*input)[0] == 'e' {
			*input = (*input)[1:]
			return list, nil
		}

		val, err := decode(input, depth)
		if err != nil {
			return nil, err
		}

		list = append(list, val)
	}
}

func decodeDict(input *string, depth *int) (map[string]any, error) {
	*input = (*input)[1:]
	dict := make(map[string]any)

	previousKey := ""
	first := true

	for {
		if len(*input) == 0 {
			return nil, MissingDictTermErr
		}
		if (*input)[0] == 'e' {
			*input = (*input)[1:]
			return dict, nil
		}

		keyNode, err := decode(input, depth)
		if err != nil {
			return nil, err
		}

		key, ok := keyNode.(string)
		if !ok {
			return nil, NonStrKeyErr
		}
		if key < previousKey && !first {
			return nil, NonSortedKeysErr
		}

		val, err := decode(input, depth)
		if err != nil {
			return nil, err
		}

		_, exists := dict[string(key)]
		if exists {
			return nil, DuplicateKeyErr
		}
		dict[string(key)] = val
		previousKey = string(key)
		first = false
	}
}
