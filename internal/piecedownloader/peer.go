package piecedownloader

type Peer interface {
	Request(idx, begin, length uint32)
	Cancel(idx, begin, length uint32)
}
