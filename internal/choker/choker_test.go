package choker_test

import (
	"Naverno/internal/choker"
	"slices"
	"testing"
	"time"
)

func TestChoker(t *testing.T) {
	p1 := choker.NewMockPeer(15, 15, true, time.Now())
	p2 := choker.NewMockPeer(20, 20, false, time.Now())
	p3 := choker.NewMockPeer(5, 5, false, time.Now())
	p4 := choker.NewMockPeer(10, 10, true, time.Now())
	peers := []choker.Peer{p1, p2, p3, p4}

	c := choker.New(time.Millisecond*1, time.Millisecond*2)
	result := c.PickPeers(peers)
	if len(result) != 3 {
		t.Fatalf("should have picked %v peers, picked %v peers instead", 3, len(result))
	}
	if slices.Contains(result, choker.Peer(p3)) {
		t.Error("peer shouldn't have been picked")
	}

	optimistic := c.PickOptimistic(peers)
	if optimistic != p3 {
		t.Error("peer should have been picked as optimistic")
	}

	c.OnInterested(p2)
	if !slices.Contains(c.Downloaders(), choker.Peer(p2)) {
		t.Error("peer should have been promoted to downloader")
	}
}
