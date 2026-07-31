package rarestfirstpicker

import (
	"Naverno/internal/picker"
	"cmp"
	"math/rand/v2"
	"slices"
)

type RarestFirstPicker struct {
	pieces       []picker.PieceState
	availability []uint32
}

func New(pieces uint32) *RarestFirstPicker {
	pickerPieces := []picker.PieceState{}
	availability := make([]uint32, pieces)
	for range pieces {
		pickerPieces = append(pickerPieces, picker.PIECE_FREE)
	}
	return &RarestFirstPicker{pieces: pickerPieces, availability: availability}
}

func (p *RarestFirstPicker) Pick(pe picker.Peer) (uint32, bool) {
	pickable := []uint32{}
	rarest := []uint32{}
	for set := range pe.GetPieces().EachSet() {
		if p.pieces[set] == picker.PIECE_FREE {
			pickable = append(pickable, uint32(set))
		}
	}

	if len(pickable) == 0 {
		return 0, false
	}

	slices.SortFunc(pickable, func(e1, e2 uint32) int { return cmp.Compare(p.availability[e1], p.availability[e2]) })

	previusAvailabilty := p.availability[0]
	for _, idx := range pickable {
		if p.availability[idx] > previusAvailabilty {
			break
		}
		rarest = append(rarest, idx)
		previusAvailabilty = p.availability[idx]
	}

	if len(rarest) > 1 {
		rand.Shuffle(len(rarest), func(i, j int) { rarest[i], rarest[j] = rarest[j], rarest[i] })
	}
	return rarest[0], true
}

func (p *RarestFirstPicker) OnPeerHave(idx uint32) {
	p.availability[idx]++
}

func (p *RarestFirstPicker) OnPeerBitfield(pe picker.Peer) {
	for set := range pe.GetPieces().EachSet() {
		p.availability[set]++
	}
}

func (p *RarestFirstPicker) OnPeerDisconnected(pe picker.Peer) {
	for set := range pe.GetPieces().EachSet() {
		if p.availability[set] > 0 {
			p.availability[set]--
		}
	}
}

func (p *RarestFirstPicker) SetFree(idx uint32) {
	p.pieces[idx] = picker.PIECE_FREE
}

func (p *RarestFirstPicker) OnPieceCompleted(idx uint32) {
	p.pieces[idx] = picker.PIECE_COMPLETED
}
