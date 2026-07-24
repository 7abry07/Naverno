package writer_test

import (
	"Naverno/internal/peer/writer"
	"Naverno/internal/peerprotocol"
	"Naverno/internal/test"
	"io"
	"log/slog"
	"testing"
	"time"
)

func TestWrite(t *testing.T) {
	conn := test.NewMockConn([]byte{})
	w := writer.New(slog.New(slog.NewTextHandler(io.Discard, nil)), conn)
	go w.Run()
	defer w.Close()

	w.Write(peerprotocol.Choke{})
	buf := make([]byte, 5)
	ok := conn.ReadSent(buf, time.Second*2)
	if !ok {
		t.Fatal("test time exceeded")
	}

	mess, err := peerprotocol.Decode(buf)
	if err != nil {
		t.Fatalf("unexpected error -> %v", err)
	}
	if mess.ID().String() != peerprotocol.MessageID.String(peerprotocol.ChokeID) {
		t.Errorf("expected -> %v, got -> %v", peerprotocol.MessageID.String(peerprotocol.ChokeID), mess.ID().String())
	}
}
