package trackertest

import (
	"Naverno/internal/tracker"
	"context"
	"net/netip"
	"time"
)

type MockWorkingTracker struct {
	announce string
}

func NewWorkingMock() *MockWorkingTracker {
	return &MockWorkingTracker{"working"}
}

func (t *MockWorkingTracker) Announce(ctx context.Context, req tracker.AnnounceRequest) (*tracker.AnnounceResponse, error) {
	p1 := netip.AddrPortFrom(netip.AddrFrom4([4]byte{192, 168, 1, 1}), 6881)
	p2 := netip.AddrPortFrom(netip.AddrFrom4([4]byte{192, 168, 1, 2}), 6881)
	p3 := netip.AddrPortFrom(netip.AddrFrom4([4]byte{192, 168, 1, 3}), 6881)

	return &tracker.AnnounceResponse{
		Interval: time.Minute * 30,
		Peers:    []netip.AddrPort{p1, p2, p3},
	}, nil
}

func (t *MockWorkingTracker) URL() string {
	return t.announce
}
