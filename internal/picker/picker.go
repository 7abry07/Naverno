package picker

import (
	"Naverno/internal/bitfield"
	"cmp"
	"math/rand/v2"
	"slices"
)

type Picker struct {
	pieces       []PieceState
	availability []uint32
}

func New(pieces uint32) *Picker {
	pickerPieces := []PieceState{}
	availability := make([]uint32, pieces)
	for range pieces {
		pickerPieces = append(pickerPieces, PIECE_FREE)
	}
	return &Picker{pieces: pickerPieces, availability: availability}
}

func (p *Picker) Pick(policy policy, peerPieces *bitfield.Bitfield) (uint32, bool) {
	switch policy {
	case RAREST_FIRST:
		return p.pickRarest(peerPieces)
	case SEQUENTIAL:
		return p.pickNext(peerPieces)
	default:
		return 0, false
	}
}

func (p *Picker) OnPeerHave(idx uint32) {
	p.availability[idx]++
}

func (p *Picker) OnPeerBitfield(pieces *bitfield.Bitfield) {
	for set := range pieces.EachSet() {
		p.availability[set]++
	}
}

func (p *Picker) OnPeerDisconnected(pieces *bitfield.Bitfield) {
	if pieces == nil {
		return
	}
	for set := range pieces.EachSet() {
		if p.availability[set] > 0 {
			p.availability[set]--
		}
	}
}

func (p *Picker) SetFree(idx uint32) {
	p.pieces[idx] = PIECE_FREE
}

func (p *Picker) OnPieceCompleted(idx uint32) {
	p.pieces[idx] = PIECE_COMPLETED
}

func (p *Picker) pickNext(peerPieces *bitfield.Bitfield) (uint32, bool) {
	for i := range peerPieces.EachSet() {
		if p.pieces[i] == PIECE_FREE {
			p.pieces[i] = PIECE_DOWNLOADING
			return uint32(i), true
		}
	}
	return 0, false
}

func (p *Picker) pickRarest(peerPieces *bitfield.Bitfield) (uint32, bool) {
	pickable := []uint32{}
	rarest := []uint32{}
	for set := range peerPieces.EachSet() {
		if p.pieces[set] == PIECE_FREE {
			pickable = append(pickable, uint32(set))
		}
	}

	if len(pickable) == 0 {
		return 0, false
	}

	slices.SortFunc(pickable, func(e1, e2 uint32) int { return cmp.Compare(p.availability[e1], p.availability[e2]) })

	previusAvailabilty := p.availability[pickable[0]]
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
	p.pieces[rarest[0]] = PIECE_DOWNLOADING
	return rarest[0], true
}
