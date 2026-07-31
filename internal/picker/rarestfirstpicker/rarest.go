package rarestfirstpicker

import (
	"Naverno/internal/picker"
	"Naverno/internal/piece"
	"cmp"
	"math/rand/v2"
	"slices"
)

type RarestFirstPicker struct {
	pieces []Piece
}

func New(pieces []*piece.Piece) *RarestFirstPicker {
	pickerPieces := []Piece{}
	for _, p := range pieces {
		pickerPieces = append(pickerPieces, Piece{Piece: picker.Piece{Piece: p, State: picker.PIECE_FREE}, availability: 0})
	}
	return &RarestFirstPicker{pieces: pickerPieces}
}

func (p *RarestFirstPicker) Pick(pe picker.Peer) *piece.Piece {
	pickable := []Piece{}
	for set := range pe.GetPieces().EachSet() {
		if p.pieces[set].State == picker.PIECE_FREE {
			pickable = append(pickable, p.pieces[set])
		}
	}
	slices.SortFunc(pickable, func(e1, e2 Piece) int { return cmp.Compare(e1.availability, e2.availability) })
	rarest := []Piece{}
	previusAvailabilty := p.pieces[0].availability
	for _, p := range pickable {
		if p.availability < previusAvailabilty {
			break
		}
		rarest = append(rarest, p)
		previusAvailabilty = p.availability
	}

	if len(rarest) > 1 {
		rand.Shuffle(len(rarest), func(i, j int) { rarest[i], rarest[j] = rarest[j], rarest[i] })
	}
	return rarest[0].Piece.Piece
}
func (p *RarestFirstPicker) OnPeerHave(pi *piece.Piece) {
	p.pieces[pi.Idx].availability++
}
func (p *RarestFirstPicker) OnPeerBitfield(pe picker.Peer) {
	for set := range pe.GetPieces().EachSet() {
		p.pieces[set].availability++
	}
}
func (p *RarestFirstPicker) OnPeerDisconnected(pe picker.Peer) {
	for set := range pe.GetPieces().EachSet() {
		p.pieces[set].availability++
	}
}
func (p *RarestFirstPicker) SetFree(pi *piece.Piece) {
	p.pieces[pi.Idx].State = picker.PIECE_FREE
}
func (p *RarestFirstPicker) OnPieceCompleted(pi *piece.Piece) {
	p.pieces[pi.Idx].State = picker.PIECE_COMPLETED
}
