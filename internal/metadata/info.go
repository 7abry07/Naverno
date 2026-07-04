package metadata

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/zeebo/bencode"
)

type File struct {
	Length int64
	Path   string
}

type file struct {
	Length int      `bencode:"length"`
	Path   []string `bencode:"path"`
}

type Info struct {
	Name        string
	Infohash    [20]byte
	PieceLength int64
	Private     bool
	Pieces      []byte
	Files       []File
}

func newInfo(in []byte) (*Info, error) {
	info := Info{}

	var infoType struct {
		Name        string `bencode:"name"`
		PieceLength int64  `bencode:"piece length"`
		Private     bool   `bencode:"private"`
		Pieces      string `bencode:"pieces"`
		Length      int64  `bencode:"length"`
		Files       []file `bencode:"files"`
	}

	decoder := bencode.NewDecoder(bytes.NewReader(in))
	decoder.SetFailOnUnorderedKeys(true)

	err := decoder.Decode(&infoType)
	if err != nil {
		return nil, err
	}

	if infoType.PieceLength == 0 {
		return nil, InvalidPieceLengthErr
	}

	if len(infoType.Pieces)%sha1.Size != 0 {
		return nil, InvalidPiecesErr
	}

	if len(infoType.Pieces)/sha1.Size == 0 {
		return nil, NoPiecesErr
	}

	info.Pieces = []byte(infoType.Pieces)
	info.Private = infoType.Private
	info.PieceLength = infoType.PieceLength
	info.Name = infoType.Name

	for _, file := range infoType.Files {
		for _, path := range file.Path {
			if strings.TrimSpace(path) == ".." {
				return nil, fmt.Errorf("invalid file name: %q", filepath.Join(file.Path...))
			}
		}
	}

	if len(infoType.Files) > 0 {
		for _, file := range infoType.Files {
			info.Files = append(info.Files, File{int64(file.Length), strings.Join(file.Path, "/")})
		}
	} else {
		info.Files = append(info.Files, File{infoType.Length, infoType.Name})
	}

	hash := sha1.New()
	hash.Write(in)
	info.Infohash = [20]byte(hash.Sum(nil))

	return &info, nil
}
