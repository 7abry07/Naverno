package announcer_test

import (
	"Naverno/internal/announcer"
	"Naverno/internal/tracker"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/netip"
	"testing"
	"time"
)

type MockFailingTracker struct {
	announce string
}

func NewFailingMock() *MockFailingTracker {
	return &MockFailingTracker{"failing"}
}

func (t *MockFailingTracker) Announce(ctx context.Context, req tracker.AnnounceRequest) (*tracker.AnnounceResponse, error) {
	return nil, fmt.Errorf("supposed to fail")
}

func (t *MockFailingTracker) URL() string {
	return t.announce
}

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

func TestAnnouncer(t *testing.T) {
	tier1 := []tracker.Tracker{NewFailingMock(), NewFailingMock(), NewFailingMock()}
	tier2 := []tracker.Tracker{NewFailingMock(), NewWorkingMock(), NewFailingMock()}
	tiers := [][]tracker.Tracker{}
	tiers = append(tiers, tier1)
	tiers = append(tiers, tier2)

	a := announcer.New(slog.New(slog.NewTextHandler(io.Discard, nil)), tiers, 6881)

	torrentC := make(chan announcer.Torrent)
	peers := make(chan []netip.AddrPort)

	go a.Run(torrentC, peers)

	testTimer := time.NewTimer(time.Second * 4)
	for {
		exit := false
		select {
		case <-torrentC:
			torrentC <- announcer.Torrent{}
		case <-peers:
			exit = true
		case <-testTimer.C:
			t.Fatal("test time exceeded")
		}
		if exit {
			break
		}
	}
}
