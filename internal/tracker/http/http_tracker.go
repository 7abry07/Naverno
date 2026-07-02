package http_tracker

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

// --------------- Structs -------------------

type HTTPTracker struct {
	announce *url.URL
	client   *http.Client
	logger   *slog.Logger

	trackerid string
}

// --------------- Functions -------------------

func New(logger *slog.Logger, url *url.URL, transport *http.Transport) *HTTPTracker {
	t := HTTPTracker{}
	t.announce = url
	t.client = &http.Client{Transport: transport}
	t.trackerid = ""
	t.logger = logger

	return &t
}

// --------------- Methods -------------------

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

	peers := []tracker.Peer{}

	if resp.peers.Type() == bencode.List_t {
		peersList, _ := resp.peers.List()
		parsedPeers, ok := tracker.ParseV4BencodedPeers(peersList)
		if !ok {
			return nil, tracker.InvalidRespErr
		}
		peers = append(peers, parsedPeers...)
	} else if resp.peers.Type() == bencode.Str_t {
		peersStr, _ := resp.peers.Str()
		parsedPeers, ok := tracker.ParseV4CompactPeers(string(peersStr))
		if !ok {
			return nil, tracker.InvalidRespErr
		}
		peers = append(peers, parsedPeers...)
	}

	if resp.peers6.Type() == bencode.Str_t {
		peersStr, _ := resp.peers6.Str()
		if len(peersStr) != 0 {
			parsedPeers, ok := tracker.ParseV6CompactPeers(string(peersStr))
			if ok {
				peers = append(peers, parsedPeers...)
			}
		}
	}

	t.logger.Debug("got peers from "+t.URL(), "peers", len(peers))

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

func (t *HTTPTracker) serialize(r tracker.AnnounceRequest) *url.URL {
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
		t.logger.Debug("want peers", "numwant", r.Numwant)
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
		return r, false
	}
	root, ok := decoded.Dict()
	if !ok {
		return r, false
	}

	interval, _ := root.FindIntOrDef("interval", 1800)
	minInterval, _ := root.FindIntOrDef("min interval", 30)
	r.interval = int32(interval)
	r.minInterval = int32(minInterval)

	warning, _ := root.FindStrOrDef("warning message", "")
	r.warning = string(warning)

	failure, failureOk := root.FindStrOrDef("failure reason", "")
	r.failure = string(failure)
	if failureOk {
		retryIn, _ := root.FindStrOrDef("retry in", "never")
		r.retryIn = string(retryIn)
		return r, true
	}

	complete, _ := root.FindIntOrDef("complete", -1)
	incomplete, _ := root.FindIntOrDef("incomplete", -1)
	r.complete = int64(complete)
	r.incomplete = int64(incomplete)

	trackerid, ok := root.FindStrOrDef("tracker id", "")
	t.trackerid = string(trackerid)

	if ok {
		t.logger.Debug("tracker id", "id", string(trackerid))
	}

	peersNode, ok := root.Find("peers")
	if !ok {
		return r, false
	}

	peers6Node, ok := root.Find("peers6")
	r.peers6 = bencode.NewStr("")
	if ok {
		r.peers6 = peers6Node
	}
	r.peers = peersNode

	return r, true
}
