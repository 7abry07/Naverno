package bencode

import (
	"fmt"
	"sort"
	"strings"
)

func Encode(n any) string {
	var result strings.Builder
	switch n := n.(type) {

	case int64:
		result.WriteString(encodeInt(n))
	case string:
		result.WriteString(encodeStr(n))
	case []any:
		result.WriteString(encodeList(n))
	case map[string]any:
		result.WriteString((encodeDict(n)))

	}
	return result.String()
}

func encodeInt(i int64) string {
	return fmt.Sprintf("i%de", i)
}

func encodeStr(s string) string {
	return fmt.Sprintf("%d:%s", len(s), s)
}

func encodeList(l []any) string {
	var result strings.Builder
	result.WriteRune('l')
	for _, val := range l {
		result.WriteString(Encode(val))
	}
	result.WriteRune('e')
	return result.String()
}

func encodeDict(d map[string]any) string {
	var result strings.Builder
	result.WriteRune('d')

	keys := make([]string, 0, len(d))
	for k := range d {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		result.WriteString(encodeStr(k))
		result.WriteString(Encode(d[k]))
	}
	result.WriteRune('e')
	return result.String()
}
