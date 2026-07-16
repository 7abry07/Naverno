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
	KeepAliveID    = 255
)

var messageStr = map[MessageID]string{
	ChokeID:        "choke",
	UnchokeID:      "unchoke",
	InterestedID:   "interested",
	UninterestedID: "uninterested",
	HaveID:         "have",
	BitfieldID:     "bitfield",
	RequestID:      "request",
	PieceID:        "piece",
	CancelID:       "cancel",
	KeepAliveID:    "keepalive",
}

func (id MessageID) String() string {
	return messageStr[id]
}

func (KeepAlive) ID() MessageID    { return KeepAliveID }
func (Choke) ID() MessageID        { return ChokeID }
func (Unchoke) ID() MessageID      { return UnchokeID }
func (Interested) ID() MessageID   { return InterestedID }
func (Uninterested) ID() MessageID { return UninterestedID }
func (Have) ID() MessageID         { return HaveID }
func (Bitfield) ID() MessageID     { return BitfieldID }
func (Request) ID() MessageID      { return RequestID }
func (Piece) ID() MessageID        { return PieceID }
func (Cancel) ID() MessageID       { return CancelID }
