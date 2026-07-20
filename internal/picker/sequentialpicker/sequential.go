package sequentialpicker

import (
	"Naverno/internal/picker"
	// "encoding/binary"
	// "fmt"
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
		//
		// if i > uint(len(p.pieces))-1 {
		// 	words := pe.GetPieces().Words()
		// 	bits := 0
		// 	for _, w := range words {
		// 		b := make([]byte, 8)
		// 		binary.BigEndian.PutUint64(b, w)
		// 		for _, by := range b {
		// 			fmt.Printf("%b", by)
		// 			bits += 8
		// 		}
		// 	}
		// 	fmt.Printf("\n")
		// 	words = pe.GetPieces().SetAll().Words()
		// 	bits = 0
		// 	for _, w := range words {
		// 		b := make([]byte, 8)
		// 		binary.BigEndian.PutUint64(b, w)
		// 		for _, by := range b {
		// 			fmt.Printf("%b", by)
		// 			bits += 8
		// 		}
		// 	}
		// 	fmt.Printf("\nall set-> %v", pe.GetPieces().All())
		// 	fmt.Printf("\nbits -> %v", bits)
		// 	fmt.Printf("\npeer pieces -> %v, peer set pieces -> %v, index -> %v\n", pe.GetPieces().Len(), pe.GetPieces().Count(), i)
		// 	fmt.Println(pe.GetPieces().String())
		// 	panic("picked out of bounds piece")
		// }
		//
		if p.pieces[uint32(i)] == picker.PIECE_FREE {
			p.pieces[uint32(i)] = picker.PIECE_DOWNLOADING
			return uint32(i), true
		}
	}
	return 0, false
}
func (p *SequentialPicker) OnPeerHave(idx uint32)             {}
func (p *SequentialPicker) OnPeerBitfield(pe picker.Peer)     {}
func (p *SequentialPicker) OnPeerDisconnected(pe picker.Peer) {}
func (p *SequentialPicker) OnPieceCompleted(idx uint32) {
	p.pieces[uint32(idx)] = picker.PIECE_COMPLETED
}
