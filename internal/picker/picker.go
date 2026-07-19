package picker

type Picker interface {
	Pick(pe Peer) (uint32, bool)
	OnPeerHave(idx uint32)
	OnPeerBitfield(pe Peer)
	OnPeerDisconnected(pe Peer)
	OnPieceCompleted(idx uint32)
}
