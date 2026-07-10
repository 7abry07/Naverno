package peerprotocol

type MessageID uint8

const (
	ChokeID        = 0
	UnchokeID      = 1
	InterestedID   = 2
	UninterestedID = 3
	HaveID         = 4
	BitfieldID     = 5
	RequestID      = 6
	PieceID        = 7
	CancelID       = 8
)
