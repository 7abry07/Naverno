package metadata

import (
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/zeebo/bencode"
)

type Metadata struct {
	Info

	AnnounceList [][]url.URL

	CreationDate time.Time
	Comment      string
	Created_by   string
}

func New(in io.Reader) (*Metadata, error) {
	meta := Metadata{}

	var root struct {
		Info         bencode.RawMessage `bencode:"info"`
		Announce     bencode.RawMessage `bencode:"announce"`
		AnnounceList bencode.RawMessage `bencode:"announce-list"`
		CreationDate int64              `bencode:"creation date"`
		Comment      string             `bencode:"comment"`
		Created_by   string             `bencode:"created by"`
	}

	decoder := bencode.NewDecoder(in)
	decoder.SetFailOnUnorderedKeys(true)
	err := decoder.Decode(&root)
	if err != nil {
		return nil, err
	}

	if len(root.Info) == 0 {
		return nil, fmt.Errorf("missing info key")
	}

	if len(root.AnnounceList) > 0 {
		al := [][]string{}
		if err := bencode.DecodeBytes(root.AnnounceList, &al); err != nil {
			return nil, err
		}

		for _, tier := range al {
			tierVal := []url.URL{}
			for _, tracker := range tier {
				parsed, err := url.Parse(tracker)
				if err != nil {
					return nil, fmt.Errorf("announce-list contains an invalid URL")
				}
				tierVal = append(tierVal, *parsed)
			}
			meta.AnnounceList = append(meta.AnnounceList, tierVal)
		}
	} else if len(root.Announce) > 0 {
		ann := ""
		if err := bencode.DecodeBytes(root.Announce, &ann); err != nil {
			return nil, err
		}

		parsed, err := url.Parse(ann)
		if err != nil {
			return nil, fmt.Errorf("announce URL is invalid")
		}
		meta.AnnounceList = append(meta.AnnounceList, []url.URL{*parsed})
	} else {
		return nil, fmt.Errorf("missing announce key")
	}

	info, err := newInfo(root.Info)

	if err != nil {
		return nil, err
	}
	meta.Info = *info

	return &meta, nil
}
