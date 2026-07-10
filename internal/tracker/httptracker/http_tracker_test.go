package httptracker_test

import (
	"Naverno/internal/tracker"
	"Naverno/internal/tracker/httptracker"
	"Naverno/internal/tracker/testserver"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/netip"
	"net/url"
	"slices"
	"testing"
)

func TestTracker(t *testing.T) {
	testserver.StartHttp()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	announce, _ := url.Parse("http://localhost:8000/announce")

	tr := httptracker.New(logger, *announce, &http.Transport{})

	peer1, _ := netip.ParseAddrPort("192.168.1.1:6881")
	peer2, _ := netip.ParseAddrPort("192.168.1.2:6882")
	peer3, _ := netip.ParseAddrPort("192.168.1.3:6883")
	req := tracker.AnnounceRequest{
		Infohash: [20]byte{},
		PeerID:   [20]byte{},
		Port:     6881,
		Numwant:  50,
		Event:    tracker.TRACKER_STARTED,
	}

	req.Ip = peer1.Addr()
	req.Port = peer1.Port()
	res, err := tr.Announce(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected err -> %v", err)
	}
	if res.Interval.Seconds() != 1800 {
		t.Error("interval is not 30 minutes")
	}
	if len(res.Peers) != 0 {
		t.Fatal("unexpected peers length")
	}

	req.Ip = peer2.Addr()
	req.Port = peer2.Port()
	res, err = tr.Announce(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected err -> %v", err)
	}
	if res.Interval.Seconds() != 1800 {
		t.Errorf("interval is not 30 minutes")
	}
	if len(res.Peers) != 1 {
		t.Fatal("unexpected peers length")
	}
	if !slices.Contains(res.Peers, peer1) {
		t.Error("peer that should be in the response isn't")
	}

	req.Ip = peer3.Addr()
	req.Port = peer3.Port()
	res, err = tr.Announce(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected err -> %v", err)
	}
	if res.Interval.Seconds() != 1800 {
		t.Errorf("interval is not 30 minutes")
	}
	if len(res.Peers) != 2 {
		t.Fatal("unexpected peers length")
	}
	if !(slices.Contains(res.Peers, peer1) && slices.Contains(res.Peers, peer2)) {
		t.Error("peers that should be in the response aren't")
	}

	req.Ip = peer1.Addr()
	req.Port = peer1.Port()
	req.Event = tracker.TRACKER_STOPPED
	res, err = tr.Announce(context.Background(), req)

	req.Ip = peer3.Addr()
	req.Port = peer3.Port()
	req.Event = tracker.TRACKER_NONE
	res, err = tr.Announce(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected err -> %v", err)
	}
	if res.Interval.Seconds() != 1800 {
		t.Errorf("interval is not 30 minutes")
	}
	if len(res.Peers) != 1 {
		t.Fatal("unexpected peers length")
	}
	if slices.Contains(res.Peers, peer1) {
		t.Error("peers that shouldn't be in the response are")
	}
}
