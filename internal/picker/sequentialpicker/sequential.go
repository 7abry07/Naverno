package sequentialpicker

import (
	"Naverno/internal/picker"
	"Naverno/internal/piece"
)

type SequentialPicker struct {
	pieces []picker.Piece
}

func NewSequentialPicker(pieces []*piece.Piece) *SequentialPicker {
	pickerPieces := []picker.Piece{}
	for _, p := range pieces {
		pickerPieces = append(pickerPieces, picker.Piece{Piece: p, State: picker.PIECE_FREE})
	}
	return &SequentialPicker{pieces: pickerPieces}
}

func (p *SequentialPicker) Pick(pe picker.Peer) *piece.Piece {
	for i := range pe.GetPieces().SetBits() {
		if p.pieces[i].State == picker.PIECE_FREE {
			p.pieces[i].State = picker.PIECE_DOWNLOADING
			return p.pieces[i].Piece
		}
	}
	return nil
}
func (p *SequentialPicker) OnPeerHave(pi *piece.Piece)        {}
func (p *SequentialPicker) OnPeerBitfield(pe picker.Peer)     {}
func (p *SequentialPicker) OnPeerDisconnected(pe picker.Peer) {}
func (p *SequentialPicker) SetFree(pi *piece.Piece) {
	p.pieces[pi.Idx].State = picker.PIECE_FREE
}
func (p *SequentialPicker) OnPieceCompleted(pi *piece.Piece) {
	p.pieces[pi.Idx].State = picker.PIECE_COMPLETED
}
