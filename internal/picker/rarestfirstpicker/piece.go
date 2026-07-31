package rarestfirstpicker

import "Naverno/internal/picker"

type Piece struct {
	picker.Piece
	availability uint32
}
