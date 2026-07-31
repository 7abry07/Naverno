package picker

import "Naverno/internal/bitfield"

type Picker interface {
	Pick(peerPieces *bitfield.Bitfield) (uint32, bool)

	OnPieceCompleted(idx uint32)
	OnPeerHave(idx uint32)
	SetFree(idx uint32)

	OnPeerBitfield(pieces *bitfield.Bitfield)
	OnPeerDisconnected(pieces *bitfield.Bitfield)
}
