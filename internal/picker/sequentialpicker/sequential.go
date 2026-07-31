package sequentialpicker

import (
	"Naverno/internal/bitfield"
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

func (p *SequentialPicker) Pick(peerPieces *bitfield.Bitfield) (uint32, bool) {
	for i := range peerPieces.EachSet() {
		if p.pieces[i] == picker.PIECE_FREE {
			p.pieces[i] = picker.PIECE_DOWNLOADING
			return uint32(i), true
		}
	}
	return 0, false
}
func (p *SequentialPicker) OnPeerHave(idx uint32)                        {}
func (p *SequentialPicker) OnPeerBitfield(pieces *bitfield.Bitfield)     {}
func (p *SequentialPicker) OnPeerDisconnected(pieces *bitfield.Bitfield) {}
func (p *SequentialPicker) SetFree(idx uint32) {
	p.pieces[idx] = picker.PIECE_FREE
}
func (p *SequentialPicker) OnPieceCompleted(idx uint32) {
	p.pieces[idx] = picker.PIECE_COMPLETED
}
