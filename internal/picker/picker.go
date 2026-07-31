package picker

type Picker interface {
	Pick(pe Peer) (uint32, bool)

	OnPieceCompleted(idx uint32)
	OnPeerHave(idx uint32)
	SetFree(idx uint32)

	OnPeerBitfield(pe Peer)
	OnPeerDisconnected(pe Peer)
}
