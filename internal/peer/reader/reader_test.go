package reader_test

import (
	"Naverno/internal/peer/reader"
	"Naverno/internal/peerprotocol"
	"Naverno/internal/test"
	"io"
	"log/slog"
	"testing"
	"time"
)

func TestRead(t *testing.T) {
	buf := []byte{}
	buf = append(buf, peerprotocol.Choke{}.Marshal()...)
	conn := test.NewMockConn(buf)

	r := reader.New(slog.New(slog.NewTextHandler(io.Discard, nil)), conn)
	go r.Run()
	defer r.Close()

	testTime := time.NewTimer(time.Second * 2)
	select {
	case mess := <-r.Messages():
		if mess.ID() != peerprotocol.ChokeID {
			t.Errorf("expected message -> %v, got -> %v", peerprotocol.MessageID.String(peerprotocol.ChokeID), mess.ID().String())
		}
	case <-testTime.C:
		t.Fatal("test time exceeded")
	}
}
