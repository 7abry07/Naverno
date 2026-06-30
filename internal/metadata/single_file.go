package metadata

// --------------- Structs -------------------

type SingleFile struct {
	name string

	pieceLength int
	pieces      []byte
	private     *bool

	length int
}

// --------------- Methods --------------------

func (s SingleFile) Name() string {
	return s.name
}

func (s SingleFile) PieceLength() int {
	return s.pieceLength
}

func (s SingleFile) Pieces() []byte {
	return s.pieces
}

func (s SingleFile) Piece(idx int) ([20]byte, bool) {
	if len(s.pieces)/20 < idx+1 {
		return [20]byte{}, false
	} else {
		return ([20]byte)(s.pieces[idx*20 : 20]), true
	}
}

func (s SingleFile) Private() (bool, bool) {
	if s.private == nil {
		return false, false
	} else {
		return *s.private, true
	}
}

func (s SingleFile) Length() int {
	return s.length
}
