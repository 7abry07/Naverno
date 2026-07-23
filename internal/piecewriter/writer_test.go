package piecewriter_test

import (
	"Naverno/internal/piece"
	"Naverno/internal/piecewriter"
	"Naverno/internal/storage/storagetest"
	"testing"
	"time"
)

func TestWriter(t *testing.T) {
	s := storagetest.NewMockStorage()
	p := piece.NewPiece(4, 10, 30, [20]byte{})
	w := piecewriter.New(p, 10, s, make([]byte, 10))
	res := make(chan *piecewriter.PieceWriter)

	go w.Run(res)

	testTime := time.NewTimer(time.Second * 2)
	select {
	case result := <-res:
		if result.Err != nil {
			t.Errorf("unexpected error -> %v", result.Err)
		}
	case <-testTime.C:
		t.Fatal("excedeed test time limit")
	}
}
