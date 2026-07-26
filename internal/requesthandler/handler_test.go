package requesthandler_test

import (
	"Naverno/internal/peerprotocol"
	"Naverno/internal/piece"
	"Naverno/internal/requesthandler"
	"Naverno/internal/storage"
	"testing"
	"time"
)

func TestHandler(t *testing.T) {
	s := storage.NewMockStorage()
	p := piece.NewPiece(0, 10, 0, [20]byte{})
	h := requesthandler.New(s, [20]byte{1}, p, peerprotocol.Request{Idx: p.Idx, Begin: 0, Length: 5})
	res := make(chan *requesthandler.RequestHandler)

	go h.Run(res)
	defer h.Close()

	testTime := time.NewTimer(time.Second * 2)
	select {
	case result := <-res:
		if result.Err != nil {
			t.Fatalf("unexpected error -> %v", result.Err)
		}
		if len(result.Data) != 5 {
			t.Errorf("expected data  length %v, got -> %v", 5, len(result.Data))
		}
	case <-testTime.C:
		t.Fatal("excedeed test time limit")
	}
}
