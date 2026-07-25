package choker_test

import (
	"Naverno/internal/choker"
	"slices"
	"testing"
	"time"
)

func TestChoker(t *testing.T) {
	p1 := choker.NewMockPeer(1, 1, true, time.Now())
	p2 := choker.NewMockPeer(2, 2, true, time.Now())
	p3 := choker.NewMockPeer(3, 3, true, time.Now())
	p4 := choker.NewMockPeer(4, 4, true, time.Now())
	p5 := choker.NewMockPeer(5, 5, false, time.Now())
	p6 := choker.NewMockPeer(0, 0, true, time.Now())
	peers := []choker.Peer{p1, p2, p3, p4, p5, p6}

	c := choker.New(time.Millisecond*1, time.Millisecond*2)
	result := c.PickPeers(peers)
	if len(result) != 5 {
		t.Fatalf("should have picked %v peers, picked %v peers instead", 5, len(result))
	}
	if slices.Contains(result, choker.Peer(p6)) {
		t.Error("peer shouldn't have been picked")
	}

	optimistic := c.PickOptimistic(peers)
	if optimistic != p6 {
		t.Error("peer should have been picked as optimistic")
	}

	replace := c.OnInterested(p5)
	if replace != p1 {
		t.Error("peer should have been picked as replacement")
	}
}
