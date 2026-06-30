package metadata

// --------------- Structs -------------------

type File struct {
	length int
	path   []string
}

type MultiFile struct {
	name string

	pieceLength int
	pieces      []byte
	private     *bool

	files []File
}

// --------------- Methods --------------------

func (m MultiFile) Name() string {
	return m.name
}

func (m MultiFile) PieceLength() int {
	return m.pieceLength
}

func (m MultiFile) Pieces() []byte {
	return m.pieces
}

func (m MultiFile) Piece(idx int) ([20]byte, bool) {
	if len(m.pieces)/20 < idx+1 {
		return [20]byte{}, false
	} else {
		return ([20]byte)(m.pieces[idx*20 : 20]), true
	}
}

func (m MultiFile) Private() (bool, bool) {
	if m.private == nil {
		return false, false
	} else {
		return *m.private, true
	}
}

func (m MultiFile) Files() []File {
	return m.files
}
