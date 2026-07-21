package picker

type Picker interface {
	Pick(pe Peer) (uint32, bool)
	OnPeerHave(idx uint32)
	OnPeerBitfield(pe Peer)
	OnPeerDisconnected(pe Peer)
	OnPieceStalled(idx uint32)
	OnPieceCompleted(idx uint32)
}
