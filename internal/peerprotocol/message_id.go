package peerprotocol

type MessageID uint8

const (
	ChokeID        MessageID = 0
	UnchokeID      MessageID = 1
	InterestedID   MessageID = 2
	UninterestedID MessageID = 3
	HaveID         MessageID = 4
	BitfieldID     MessageID = 5
	RequestID      MessageID = 6
	PieceID        MessageID = 7
	CancelID       MessageID = 8
	ExtendedID     MessageID = 20
	KeepAliveID    MessageID = 255
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
	ExtendedID:     "extended",
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
func (Extended) ID() MessageID     { return ExtendedID }
