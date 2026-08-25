package magnet_test

import (
	"Naverno/internal/magnet"
	"bufio"
	"strings"
	"testing"
)

func TestMagnet(t *testing.T) {
	uri := "magnet:?xt=urn:btih:6926cf30eb3357268fa6688d039f78b944f147cf&dn=hello%20world&tr=http%3A%2F%2Fnyaa.tracker.wf%3A7777%2Fannounce&x.pe=127.0.0.1:8000"

	m, err := magnet.New(bufio.NewReader(strings.NewReader(uri)))
	if err != nil {
		t.Fatalf("error -> %v", err)
	}
	name := "hello world"
	if m.Name != name {
		t.Errorf("expected: %v | got: %v", name, m.Name)
	}
	announce := "http://nyaa.tracker.wf:7777/announce"
	if len(m.Trackers) == 0 {
		t.Errorf("no trackers")
	}
	if m.Trackers[0].String() != announce {
		t.Errorf("expected: %v | got: %v", announce, m.Trackers[0].String())
	}

	peer := "127.0.0.1:8000"
	if len(m.Peers) == 0 {
		t.Errorf("no peers")
	}

	if m.Peers[0] != peer {
		t.Errorf("expected: %v | got: %v", peer, m.Peers[0])
	}
}
