package handshaker

import (
	"Naverno/internal/test"
	"bytes"
	"testing"
	"time"
)

func TestValidIncomingHandshake(t *testing.T) {
	remote := test.NewMockConn(Handshake{
		InfoHash:   [20]byte{0},
		PeerID:     [20]byte{0},
		Extensions: [8]byte{0},
	}.Marshal())

	result := make(chan *IncomingHandshaker)
	incoming := NewIncomingHandshaker(remote)
	go incoming.Run(result, func(b [20]byte) bool { return true }, [20]byte{1}, [8]byte{1}, time.Second*3)

	testTimer := time.NewTimer(time.Second * 3)
	select {
	case <-testTimer.C:
		t.Fatal("test time exceeded")
	case hs := <-result:
		if hs.Error != nil {
			t.Errorf("handshake returned error -> %v", hs.Error)
		}
	}

	buf := make([]byte, 68)
	_, err := remote.ReadSent(buf)
	if err != nil {
		t.Fatalf("unexpected error -> %v", err)
	}

	hs := Handshake{}
	err = hs.Unmarshal(bytes.NewBuffer(buf))
	if err != nil {
		t.Fatalf("unexpected error -> %v", err)
	}

	if hs.PeerID != [20]byte{1} {
		t.Errorf("expected peer id -> %#v | got -> %#v", [20]byte{1}, hs.PeerID)
	}

	if hs.Extensions != [8]byte{1} {
		t.Errorf("expected extensions -> %#v | got -> %#v", [8]byte{1}, hs.Extensions)
	}
}
