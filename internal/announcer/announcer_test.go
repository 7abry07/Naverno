package announcer_test

import (
	"Naverno/internal/announcer"
	"Naverno/internal/tracker"
	"io"
	"log/slog"
	"net/netip"
	"testing"
	"time"
)

func TestAnnouncer(t *testing.T) {
	tier1 := []tracker.Tracker{tracker.NewFailingMock(), tracker.NewFailingMock(), tracker.NewFailingMock()}
	tier2 := []tracker.Tracker{tracker.NewFailingMock(), tracker.NewWorkingMock(), tracker.NewFailingMock()}
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
