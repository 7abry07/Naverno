package choker_test

import (
	"Naverno/internal/choker"
	"slices"
	"testing"
	"time"
)

type MockPeer struct {
	Upload      uint64
	Download    uint64
	Interested  bool
	connectedAt time.Time
}

func NewMockPeer(upload, download uint64, interested bool, connectedAt time.Time) *MockPeer {
	return &MockPeer{
		Upload:      upload,
		Download:    download,
		Interested:  interested,
		connectedAt: connectedAt,
	}
}

func (p *MockPeer) UploadRate() uint64 {
	return p.Upload
}
func (p *MockPeer) DownloadRate() uint64 {
	return p.Download
}
func (p *MockPeer) ConnectedAt() time.Time {
	return p.connectedAt
}
func (p *MockPeer) IsInterested() bool {
	return p.Interested
}

func TestChoker(t *testing.T) {
	p1 := NewMockPeer(1, 1, true, time.Now())
	p2 := NewMockPeer(2, 2, true, time.Now())
	p3 := NewMockPeer(3, 3, true, time.Now())
	p4 := NewMockPeer(4, 4, true, time.Now())
	p5 := NewMockPeer(5, 5, false, time.Now())
	p6 := NewMockPeer(0, 0, true, time.Now())
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
