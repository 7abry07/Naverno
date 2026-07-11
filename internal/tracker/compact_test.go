package tracker_test

import (
	"Naverno/internal/tracker"
	"testing"
)

func TestCompactPeers(t *testing.T) {
	p1, err := tracker.NewCompactPeer("192.168.1.1", 80)
	if err != nil {
		t.Fatalf("(NewCompactPeer) unexpected error -> %v", err)
	}

	marshaled, err := p1.MarshalBinary()
	if err != nil {
		t.Fatalf("(MarshalBinary) unexpected error -> %v", err)
	}

	p2, err := tracker.NewCompactPeer("", 0)
	if err != nil {
		t.Fatalf("(NewCompactPeer) unexpected error -> %v", err)
	}

	err = p2.UnmarshalBinary(marshaled)
	if err != nil {
		t.Fatalf("(UnmarshalBinary) unexpected error -> %v", err)
	}

	if p1.Ip != p2.Ip || p1.Port != p2.Port {
		t.Errorf("expected ->  [%v:%v] | got -> [%v:%v]", p1.Ip, p1.Port, p2.Ip, p2.Port)
	}
}
