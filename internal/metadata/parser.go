package metadata

import (
	"Naverno/internal/bencode"
	"crypto/sha1"
	"net/url"
	"strings"
)

func Parse(input string) (Metadata, error) {
	m := Metadata{}

	decoded, err := bencode.Decode(input)
	if err != nil {
		return m, err
	}

	root, ok := decoded.(map[string]any)
	if !ok {
		return m, RootNotDictErr
	}

	announce, ok := root["announce"].(string)
	announce_list, ok1 := root["announce-list"].([]any)
	if !ok && !ok1 {
		return m, MissingAnnounceErr
	}

	parsed_announce, ok := parseAnnounce(announce)
	if !ok {
		return m, InvalidAnnounceErr
	}
	m.announce = parsed_announce

	parsed_announce_list, ok := parseAnnounceList(announce_list)
	if ok {
		m.announce_list = &parsed_announce_list
	}

	creation_date, ok := root["creation date"].(int64)
	if ok {
		temp := int(creation_date)
		m.creation_date = &temp
	}

	comment, ok := root["comment"].(string)
	if ok {
		temp := string(comment)
		m.comment = &temp
	}

	created_by, ok := root["created by"].(string)
	if ok {
		temp := string(created_by)
		m.created_by = &temp
	}

	encoding, ok := root["encoding"].(string)
	if ok {
		temp := string(encoding)
		m.encoding = &temp
	}

	info, ok := root["info"].(map[string]any)
	if !ok {
		return m, MissingInfoErr
	}

	hash := sha1.Sum([]byte(bencode.Encode(info)))
	m.infohash = hash

	infoMarshaled, err := parseInfo(info)
	if err != nil {
		return m, err
	}

	m.Info = infoMarshaled

	return m, nil
}

func parseInfo(info map[string]any) (Info, error) {
	piece_length, ok := info["piece length"].(int64)
	if !ok {
		return nil, MissingPieceLenErr
	}

	pieces, ok := info["pieces"].(string)
	if !ok {
		return nil, MissingPiecesErr
	}
	if len(pieces)%20 != 0 {
		return nil, InvalidPiecesErr
	}

	name, ok := info["name"].(string)
	if !ok {
		return nil, MissingNameErr
	}

	length, ok := info["length"].(int64)
	filesList, ok1 := info["files"].([]any)

	if ok && ok1 {
		return nil, BothLengthFilesPresentErr
	} else if !ok && !ok1 {
		return nil, BothLengthFilesMissingErr
	}

	if ok {
		single := SingleFile{}
		single.name = string(name)
		single.pieceLength = (int)(piece_length)
		single.pieces = ([]byte)(pieces)
		single.length = int(length)

		private, ok := info["private"].(int64)
		if ok {
			temp := int(private)
			single.private = &temp
		}
		return single, nil
	} else {
		multi := MultiFile{}
		multi.name = string(name)
		multi.pieceLength = (int)(piece_length)
		multi.pieces = ([]byte)(pieces)

		files, ok := parseFiles(filesList)
		if !ok {
			return nil, InvalidFilesErr
		}
		multi.files = files

		private, ok := info["private"].(int64)
		if ok {
			temp := int(private)
			multi.private = &temp
		}

		return multi, nil
	}
}

func parseAnnounce(announce string) (*url.URL, bool) {
	parsed, err := url.Parse(string(announce))
	if err != nil {
		return nil, false
	}
	return parsed, true
}

func parseAnnounceList(announce_list []any) ([][]*url.URL, bool) {
	result := [][]*url.URL{}
	for _, lstnode := range announce_list {
		lst, ok := lstnode.([]any)
		if !ok {
			return nil, false
		}
		resultLst := []*url.URL{}
		for _, strnode := range lst {
			str, ok := strnode.(string)
			if !ok {
				return nil, false
			}
			ann, err := url.Parse(string(str))
			if err != nil {
				return nil, false
			}
			resultLst = append(resultLst, ann)
		}
		result = append(result, resultLst)
	}

	return result, true
}
func parseFiles(files []any) ([]File, bool) {
	result := []File{}

	for _, node := range files {
		dict, ok := node.(map[string]any)
		if !ok {
			return []File{}, false
		}
		file, ok := parseFile(dict)
		if !ok {
			return []File{}, false
		}
		result = append(result, file)
	}
	return result, true
}

func parseFile(file map[string]any) (File, bool) {
	length, lOk := file["length"].(int64)
	pathlst, pOk := file["path"].([]any)
	if !lOk || !pOk {
		return File{}, false
	}

	var path strings.Builder
	for i, frag := range pathlst {
		strval, ok := frag.(string)
		if !ok {
			return File{}, false
		}

		if i == len(pathlst)-1 {
			path.WriteString(string(strval))
		} else {
			path.WriteString(string(strval) + "/")
		}
	}

	return File{Path: path.String(), Length: int(length)}, true
}
