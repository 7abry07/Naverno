package sequentialpicker

import (
	"Naverno/internal/picker"
)

type SequentialPicker struct {
	pieces map[uint32]picker.PieceState
}

func NewSequentialPicker(pieceCount uint32) *SequentialPicker {
	pieces := make(map[uint32]picker.PieceState)
	for i := range pieceCount {
		pieces[i] = picker.PIECE_FREE
	}
	return &SequentialPicker{pieces: pieces}
}

func (p *SequentialPicker) Pick(pe picker.Peer) (uint32, bool) {
	for i := range pe.GetPieces().EachSet() {
		if p.pieces[uint32(i)] == picker.PIECE_FREE ||
			p.pieces[uint32(i)] == picker.PIECE_STALLED {
			p.pieces[uint32(i)] = picker.PIECE_DOWNLOADING
			return uint32(i), true
		}
	}
	return 0, false
}
func (p *SequentialPicker) OnPeerHave(idx uint32)             {}
func (p *SequentialPicker) OnPeerBitfield(pe picker.Peer)     {}
func (p *SequentialPicker) OnPeerDisconnected(pe picker.Peer) {}
func (p *SequentialPicker) OnPieceStalled(idx uint32) {
	p.pieces[idx] = picker.PIECE_STALLED
}
func (p *SequentialPicker) OnPieceCompleted(idx uint32) {
	p.pieces[idx] = picker.PIECE_COMPLETED
}
