package metadata

import (
	"GoBit/internal/bencode"
	"crypto/sha1"
	"net/url"
	"strings"
)

// --------------- Functions -------------------

func Parse(input string) (Metadata, error) {
	m := Metadata{}

	decoded, err := bencode.Decode(input)
	if err != nil {
		return m, err
	}

	root, ok := decoded.Dict()
	if !ok {
		return m, RootNotDictErr
	}

	announce, ok := root.FindStr("announce")
	announce_list, ok1 := root.FindList("announce-list")
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

	creation_date, ok := root.FindInt("creation date")
	if ok {
		temp := int(creation_date)
		m.creation_date = &temp
	}

	comment, ok := root.FindStr("comment")
	if ok {
		temp := string(comment)
		m.comment = &temp
	}

	created_by, ok := root.FindStr("created by")
	if ok {
		temp := string(created_by)
		m.created_by = &temp
	}

	encoding, ok := root.FindStr("encoding")
	if ok {
		temp := string(encoding)
		m.encoding = &temp
	}

	infoNode, ok := root.Find("info")
	if !ok {
		return m, MissingInfoErr
	}
	hash := sha1.Sum([]byte(bencode.Encode(infoNode)))
	m.infohash = hash

	infoDict, ok := infoNode.Dict()
	if !ok {
		return m, MissingInfoErr
	}

	info, err := parseInfo(infoDict)
	if err != nil {
		return m, err
	}

	m.Info = info

	return m, nil
}

func parseInfo(info bencode.BDict) (Info, error) {
	piece_length, ok := info.FindInt("piece length")
	if !ok {
		return nil, MissingPieceLenErr
	}

	pieces, ok := info.FindStr("pieces")
	if !ok {
		return nil, MissingPiecesErr
	}
	if len(pieces)%20 != 0 {
		return nil, InvalidPiecesErr
	}

	name, ok := info.FindStr("name")
	if !ok {
		return nil, MissingNameErr
	}

	length, ok := info.FindInt("length")
	filesDict, ok1 := info.FindDict("files")

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

		private, ok := info.FindInt("private")
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

		files, ok := parseFiles(filesDict)
		if !ok {
			return nil, InvalidFilesErr
		}
		multi.files = files

		private, ok := info.FindInt("private")
		if ok {
			temp := int(private)
			multi.private = &temp
		}

		return multi, nil
	}
}

func parseAnnounce(announce bencode.BStr) (*url.URL, bool) {
	parsed, err := url.Parse(string(announce))
	if err != nil {
		return nil, false
	}
	return parsed, true
}

func parseAnnounceList(announce_list bencode.BList) ([][]*url.URL, bool) {
	result := [][]*url.URL{}
	for _, lstnode := range announce_list {
		lst, ok := lstnode.List()
		if !ok {
			return nil, false
		}
		resultLst := []*url.URL{}
		for _, strnode := range lst {
			str, ok := strnode.Str()
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
func parseFiles(files bencode.BDict) ([]File, bool) {
	result := []File{}

	for _, node := range files {
		dict, ok := node.Dict()
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

func parseFile(file bencode.BDict) (File, bool) {
	length, lOk := file.FindInt("length")
	pathlst, pOk := file.FindList("path")
	if !lOk || !pOk {
		return File{}, false
	}

	var path strings.Builder
	for i, frag := range pathlst {
		strval, ok := frag.Str()
		if !ok {
			return File{}, false
		}

		if i == len(pathlst)-1 {
			path.WriteString(string(strval))
		} else {
			path.WriteString(string(strval) + "/")
		}
	}

	return File{Path: path.String(), Length: uint(length)}, true
}
