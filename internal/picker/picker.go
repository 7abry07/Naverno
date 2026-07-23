package picker

import "Naverno/internal/piece"

type Picker interface {
	Pick(pe Peer) *piece.Piece

	OnPieceCompleted(p *piece.Piece)
	OnPeerHave(p *piece.Piece)
	SetFree(p *piece.Piece)

	OnPeerBitfield(pe Peer)
	OnPeerDisconnected(pe Peer)
}
