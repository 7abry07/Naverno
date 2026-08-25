package magnet

import (
	"bufio"
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"strings"
)

type Magnet struct {
	Infohash [20]byte
	Name     string
	Trackers []url.URL
	Peers    []string
}

func New(in *bufio.Reader) (*Magnet, error) {
	m := Magnet{}
	start, err := in.ReadString('?')
	if err != nil {
		return nil, fmt.Errorf("error while parsing magnet URI-> %v", err)
	}
	if string(start) != "magnet:?" {
		return nil, fmt.Errorf("not a magnet link")
	}

	ih := [20]byte{}
	name := ""
	trackers := []string{}
	peers := []string{}

	for in.Buffered() != 0 {
		key, err := in.ReadString('=')
		if err != nil {
			return nil, fmt.Errorf("error while parsing magnet URI-> %v", err)
		}
		key = key[:len(key)-1]

		value, err := in.ReadString('&')
		switch err {
		case io.EOF:
		case nil:
			value = value[:len(value)-1]
		default:
			return nil, fmt.Errorf("error while parsing magnet URI-> %v", err)
		}

		switch key {
		case "xt":
			hashType := value[:9]
			hash := value[9:]

			switch hashType {
			case "urn:btih:":
				switch len(hash) {
				case 40:
					decoded, err := hex.DecodeString(hash)
					if err != nil {
						return nil, fmt.Errorf("invalid hex encoded infohash -> %v", err)
					}
					ih = [20]byte(decoded)
				case 32:
					decoder := base32.NewDecoder(base32.StdEncoding, strings.NewReader(hash))
					_, err := decoder.Read(ih[:])
					if err != nil {
						return nil, fmt.Errorf("invalid base32 encoded infohash -> %v", err)
					}
				}
			case "urn:btmh:":
				return nil, fmt.Errorf("multi hash not supported in magnet URI")
			}
		case "dn":
			name = value
		case "tr":
			trackers = append(trackers, value)
		case "x.pe":
			peers = append(peers, value)
		}
	}

	m.Infohash = ih

	m.Name, err = url.PathUnescape(name)
	if err != nil {
		m.Name = ""
	}

	for _, tr := range trackers {
		unescaped, err := url.PathUnescape(tr)
		if err != nil {
			continue
		}
		parsed, err := url.Parse(unescaped)
		if err != nil {
			continue
		}
		m.Trackers = append(m.Trackers, *parsed)
	}

	m.Peers = peers

	return &m, nil
}
