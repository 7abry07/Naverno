package httptracker

import (
	"Naverno/internal/bencode"
	"Naverno/internal/tracker"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type HTTPTracker struct {
	announce url.URL
	client   *http.Client
	logger   *slog.Logger

	trackerid string
}

func New(logger *slog.Logger, announce url.URL, transport *http.Transport) *HTTPTracker {
	t := HTTPTracker{}
	t.announce = announce
	t.client = &http.Client{Transport: transport}
	t.trackerid = ""
	t.logger = logger

	return &t
}

func (t *HTTPTracker) Announce(ctx context.Context, r tracker.AnnounceRequest) (*tracker.AnnounceResponse, error) {
	urlReq := t.serialize(r)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", urlReq.String(), nil)
	if err != nil {
		return nil, err
	}

	httpResp, err := t.client.Do(httpReq)
	if err != nil {
		return nil, err
	}

	defer httpResp.Body.Close()
	httpBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	resp, ok := t.deserialize(httpBody)
	if !ok {
		return nil, tracker.InvalidRespErr
	}

	if resp.failure != "" {
		retryin, _ := strconv.Atoi(resp.retryIn)
		return nil, tracker.Error{
			Reason:  resp.failure,
			RetryIn: time.Duration(retryin) * time.Second,
		}
	}

	peers := []netip.AddrPort{}

	switch peersVal := resp.peers.(type) {
	case []any:
		{
			parsedPeers, ok := ParseBencodedPeers(peersVal)
			if !ok {
				return nil, tracker.InvalidRespErr
			}
			peers = append(peers, parsedPeers...)
		}
	case string:
		{
			parsedPeers, ok := tracker.ParseV4CompactPeers(peersVal)
			if !ok {
				return nil, tracker.InvalidRespErr
			}
			peers = append(peers, parsedPeers...)
		}

	}

	switch peersVal := resp.peers6.(type) {
	case string:
		{
			if len(peersVal) != 0 {
				parsedPeers, ok := tracker.ParseV6CompactPeers(peersVal)
				if ok {
					peers = append(peers, parsedPeers...)
				}
			}
		}
	}

	t.logger.Debug("got peers", "tracker", t.URL(), "peers", len(peers))

	return &tracker.AnnounceResponse{
		MinInterval:    time.Duration(resp.minInterval) * time.Second,
		Interval:       time.Duration(resp.interval) * time.Second,
		Leechers:       resp.incomplete,
		Seeders:        resp.complete,
		WarningMessage: resp.warning,
		Peers:          peers,
	}, nil
}

func (t *HTTPTracker) URL() string {
	return t.announce.String()
}

func (t *HTTPTracker) serialize(r tracker.AnnounceRequest) url.URL {
	fullUrl := t.announce

	query := strings.Builder{}

	query.WriteString("info_hash=")
	query.WriteString(url.QueryEscape(string(r.Infohash[:])))

	query.WriteString("&peer_id=")
	query.WriteString(url.QueryEscape(string(r.PeerID[:])))

	query.WriteString("&port=")
	query.WriteString(url.QueryEscape(strconv.Itoa(int(r.Port))))

	query.WriteString("&uploaded=")
	query.WriteString(url.QueryEscape(strconv.Itoa(int(r.Uploaded))))

	query.WriteString("&left=")
	query.WriteString(url.QueryEscape(strconv.Itoa(int(r.Left))))

	query.WriteString("&compact=1")
	query.WriteString("&no_peer_id=1")

	query.WriteString("&event=")
	query.WriteString(url.QueryEscape(r.Event.String()))

	query.WriteString("&key=")
	query.WriteString(url.QueryEscape(string(r.PeerID[16:20])))

	if r.Numwant != 0 {
		query.WriteString("&numwant=")
		query.WriteString(url.QueryEscape(strconv.Itoa(int(r.Numwant))))
		t.logger.Debug("want peers", "tracker", t.URL(), "numwant", r.Numwant)
	}

	if (r.Ip != netip.Addr{}) {
		query.WriteString("&ip=")
		query.WriteString(url.QueryEscape(r.Ip.String()))
	}

	if t.trackerid != "" {
		query.WriteString("&trackerid=")
		query.WriteString(url.QueryEscape(t.trackerid))
	}

	fullUrl.RawQuery = query.String()

	return fullUrl
}

func (t *HTTPTracker) deserialize(httpResp []byte) (announceResponse, bool) {
	r := announceResponse{}

	decoded, err := bencode.Decode(string(httpResp))
	if err != nil {
		t.logger.Error("error in decoding tracker response", "tracker", t.URL())
		return r, false
	}
	root, ok := decoded.(map[string]any)
	if !ok {
		t.logger.Error("the response is not a bencode dictionary", "tracker", t.URL())
		return r, false
	}

	interval, ok := root["interval"].(int64)
	if !ok {
		interval = 1800
	}
	r.interval = int32(interval)

	minInterval, ok := root["min interval"].(int64)
	if !ok {
		interval = 1800
	}
	r.minInterval = int32(minInterval)

	warning, ok := root["warning message"].(string)
	if !ok {
		warning = ""
	}
	r.warning = warning

	failure, ok := root["failure reason"].(string)
	if ok {
		r.failure = failure
		retryIn, _ := root["retry in"].(string)
		r.retryIn = retryIn
		return r, true
	}

	complete, ok := root["complete"].(int64)
	if !ok {
		complete = -1
	}
	incomplete, ok := root["incomplete"].(int64)
	if !ok {
		incomplete = -1
	}

	r.complete = int64(complete)
	r.incomplete = int64(incomplete)

	trackerid, ok := root["tracker id"].(string)
	t.trackerid = trackerid

	if ok {
		t.logger.Debug("tracker id", "id", trackerid)
	}

	peersNode, ok := root["peers"]
	if !ok {
		t.logger.Error("no peers in tracker response", "tracker", t.URL())
		return r, false
	}

	peers6Node, ok := root["peers6"]
	if ok {
		r.peers6 = peers6Node
	}
	r.peers = peersNode

	return r, true
}
