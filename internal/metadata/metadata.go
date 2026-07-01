package metadata

import (
	"net/url"
	"time"
)

// --------------- Structs -------------------

type Metadata struct {
	Info

	infohash [20]byte

	announce      *url.URL
	announce_list *[][]*url.URL

	creation_date *int
	comment       *string
	created_by    *string
	encoding      *string
}

type File struct {
	Length int
	Path   string
}

// -------------- Interfaces ------------------

type Info interface {
	Name() string

	PieceLength() int
	Pieces() []byte
	Piece(int) ([20]byte, bool)

	Private() (bool, bool)

	Files() []File
}

// --------------- Methods --------------------

func (m Metadata) Announce() *url.URL {
	return m.announce
}

func (m Metadata) Infohash() [20]byte {
	return m.infohash
}

func (m Metadata) AnnounceList() ([][]*url.URL, bool) {
	if m.announce_list == nil {
		return [][]*url.URL{}, false
	} else {
		return *m.announce_list, true
	}
}

func (m Metadata) CreationDate() (time.Time, bool) {
	if m.creation_date == nil {
		return time.Now(), false
	} else {
		return time.Unix(int64(*m.creation_date), 0), true
	}
}

func (m Metadata) CreatedBy() (string, bool) {
	if m.created_by == nil {
		return "", false
	} else {
		return *m.created_by, true
	}
}

func (m Metadata) Comment() (string, bool) {
	if m.comment == nil {
		return "", false
	} else {
		return *m.comment, true
	}
}

func (m Metadata) Encoding() (string, bool) {
	if m.encoding == nil {
		return "", false
	} else {
		return *m.encoding, true
	}
}
