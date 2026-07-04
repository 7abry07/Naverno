package httptracker

import (
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

	"github.com/zeebo/bencode"
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

	if resp.Failure != "" {
		retryin, _ := strconv.Atoi(resp.RetryIn)
		return nil, tracker.Error{
			Reason:  resp.Failure,
			RetryIn: time.Duration(retryin) * time.Second,
		}
	}

	peers := []netip.AddrPort{}

	if len(resp.Peers) > 0 {
		if resp.Peers[0] == 'l' {
			parsedPeers, ok := ParseBencodedPeers(resp.Peers)
			if !ok {
				return nil, tracker.InvalidRespErr
			}
			peers = append(peers, parsedPeers...)
		} else {

			_, val, _ := strings.Cut(string(resp.Peers), ":")
			parsedPeers, ok := tracker.ParseV4CompactPeers([]byte(val))
			if !ok {
				return nil, tracker.InvalidRespErr
			}
			peers = append(peers, parsedPeers...)
		}
	} else {
		return nil, tracker.InvalidRespErr
	}

	if len(resp.Peers6) > 0 {

		_, val, _ := strings.Cut(string(resp.Peers), ":")
		parsedPeers, ok := tracker.ParseV6CompactPeers([]byte(val))
		if ok {
			peers = append(peers, parsedPeers...)
		}
	}

	t.logger.Debug("got peers", "tracker", t.URL(), "peers", len(peers))

	return &tracker.AnnounceResponse{
		MinInterval:    time.Duration(resp.MinInterval) * time.Second,
		Interval:       time.Duration(resp.Interval) * time.Second,
		Leechers:       resp.Incomplete,
		Seeders:        resp.Complete,
		WarningMessage: resp.Warning,
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

func (t *HTTPTracker) deserialize(resp []byte) (announceResponse, bool) {
	r := announceResponse{
		Interval:    1800,
		MinInterval: 30,
		Warning:     "",
		Failure:     "",
		RetryIn:     "",
		TrackerID:   "",
		Complete:    -1,
		Incomplete:  -1,
	}

	err := bencode.DecodeBytes(resp, &r)
	if err != nil {
		t.logger.Error("error in decoding tracker response", "tracker", t.URL())
		return r, false
	}

	if r.TrackerID != "" {
		t.trackerid = r.TrackerID
		t.logger.Debug("new tracker id received", "tracker", t.URL(), "tracker id", r.TrackerID)
	}

	return r, true
}
