package tracker_test

import (
	"Naverno/internal/tracker"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestAnnounce(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		infohash := [20]byte{1}
		peerId := [20]byte{2}
		escapedInfohash := url.QueryEscape(string(infohash[:]))
		escapedPeerId := url.QueryEscape(string(peerId[:]))

		if url.QueryEscape(q.Get("info_hash")) != escapedInfohash {
			t.Errorf("expected -> [%v] | got -> [%v]", escapedInfohash, url.QueryEscape(q.Get("info_hash")))
		}

		if url.QueryEscape(q.Get("peer_id")) != escapedPeerId {
			t.Errorf("expected -> [%v] | got -> [%v]", escapedPeerId, url.QueryEscape(q.Get("peer_id")))
		}

		if q.Get("downloaded") != "23" {
			t.Errorf("expected -> [%v] | got -> [%v]", 23, q.Get("downloaded"))
		}

		if q.Get("uploaded") != "42" {
			t.Errorf("expected -> [%v] | got -> [%v]", 42, q.Get("uploaded"))
		}

		if q.Get("left") != "0" {
			t.Errorf("expected -> [%v] | got -> [%v]", 0, q.Get("left"))
		}

		if q.Get("port") != "6881" {
			t.Errorf("expected -> [%v] | got -> [%v]", 6881, q.Get("port"))
		}

		if q.Get("event") != "started" {
			t.Errorf("expected -> [%v] | got -> [%v]", "started", q.Get("event"))
		}

		switch r.URL.Path {
		case "/regular":
			w.Write([]byte("d8:completei34e8:intervali1800e5:peers0:e"))
		case "/failure":
			w.Write([]byte("d14:failure reason11:bad requeste"))
		}

	}))

	defer server.Close()

	regular := server.URL + "/regular"
	failure := server.URL + "/failure"

	regularTracker, err := tracker.NewTracker(regular)
	if err != nil {
		t.Fatalf("unexpected error -> %v", err)
	}

	req := tracker.AnnounceRequest{
		Infohash:   [20]byte{1},
		PeerID:     [20]byte{2},
		Downloaded: 23,
		Uploaded:   42,
		Left:       0,
		Port:       6881,
		Event:      tracker.TRACKER_STARTED,
	}

	// regular response

	regularResp, err := regularTracker.Announce(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error -> %v", err)
	}

	if regularResp.Complete != 34 {
		t.Errorf("expected -> [%v] | got -> [%v]", 34, regularResp.Complete)
	}

	if regularResp.Interval != 1800 {
		t.Errorf("expected -> [%v] | got -> [%v]", 1800, regularResp.Interval)
	}

	// failure response

	failureTracker, err := tracker.NewTracker(failure)
	if err != nil {
		t.Fatalf("unexpected error -> %v", err)
	}

	failureResp, err := failureTracker.Announce(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error -> %v", err)
	}

	if failureResp.Failure == nil {
		t.Errorf("expected -> [%v] | got -> [%v]", "bad request", "nil")
	}

	if *failureResp.Failure != "bad request" {
		t.Errorf("expected -> [%v] | got -> [%v]", "bad request", failureResp.Failure)
	}
}
