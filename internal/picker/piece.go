package picker

type PieceState uint8

const (
	PIECE_FREE PieceState = iota
	PIECE_STALLED
	PIECE_DOWNLOADING
	PIECE_COMPLETED
)
