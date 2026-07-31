package sequentialpicker

import (
	"Naverno/internal/picker"
)

type SequentialPicker struct {
	pieces []picker.PieceState
}

func New(pieces uint32) *SequentialPicker {
	pickerPieces := []picker.PieceState{}
	for range pieces {
		pickerPieces = append(pickerPieces, picker.PIECE_FREE)
	}
	return &SequentialPicker{pieces: pickerPieces}
}

func (p *SequentialPicker) Pick(pe picker.Peer) (uint32, bool) {
	for i := range pe.GetPieces().EachSet() {
		if p.pieces[i] == picker.PIECE_FREE {
			p.pieces[i] = picker.PIECE_DOWNLOADING
			return uint32(i), true
		}
	}
	return 0, false
}
func (p *SequentialPicker) OnPeerHave(idx uint32)             {}
func (p *SequentialPicker) OnPeerBitfield(pe picker.Peer)     {}
func (p *SequentialPicker) OnPeerDisconnected(pe picker.Peer) {}
func (p *SequentialPicker) SetFree(idx uint32) {
	p.pieces[idx] = picker.PIECE_FREE
}
func (p *SequentialPicker) OnPieceCompleted(idx uint32) {
	p.pieces[idx] = picker.PIECE_COMPLETED
}
