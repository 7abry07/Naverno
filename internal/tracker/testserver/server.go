package testserver

import (
	"Naverno/internal/tracker"
	"errors"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"sync"

	"github.com/zeebo/bencode"
)

const (
	maxNumwant       = 100
	defNumwant       = 50
	announceInterval = 1800
)

type announceResponse struct {
	Failure  string             `bencode:"failure reason,omitempty"`
	RetryIn  string             `bencode:"retry in,omitempty"`
	Interval int64              `bencode:"interval,omitempty"`
	Peers    bencode.RawMessage `bencode:"peers,omitempty"`
	Peers6   bencode.RawMessage `bencode:"peers6,omitempty"`
}

type announceRequest struct {
	InfoHash [20]byte
	PeerID   [20]byte
	Port     uint16
	Ip       netip.Addr
	Event    string
	Numwant  int64
	Compact  bool
}

type HTTPServer struct {
	store map[[20]byte][]tracker.CompactPeer
	m     sync.RWMutex
}

func StartHttp() func() error {
	s := HTTPServer{}
	s.store = make(map[[20]byte][]tracker.CompactPeer)
	s.m = sync.RWMutex{}

	listener, err := net.Listen("tcp", ":8000")
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/announce", s.announce)
	go http.Serve(listener, mux)

	return listener.Close
}

func (s *HTTPServer) announce(w http.ResponseWriter, r *http.Request) {
	s.m.Lock()
	defer s.m.Unlock()

	req, err := parseRequest(r.URL.Query())
	if err != nil {
		s.sendFailure(err.Error(), w)
	}

	port := req.Port
	ip := r.RemoteAddr
	if (req.Ip != netip.Addr{}) {
		ip = req.Ip.String()
	}
	thisPeer, _ := tracker.NewCompactPeer(ip, port)

	peers, ok := s.store[req.InfoHash]
	if !ok {
		s.store[req.InfoHash] = []tracker.CompactPeer{}
	}

	if len(peers) > int(min(maxNumwant, req.Numwant)) {
		if req.Numwant == -1 {
			peers = peers[0:defNumwant]
		} else {
			peers = peers[0:min(maxNumwant, req.Numwant)]
		}
	}

	switch req.Event {
	case "":
	case "started":
		s.store[req.InfoHash] = append(s.store[req.InfoHash], thisPeer)
	case "stopped":
		out := []tracker.CompactPeer{}

		for _, p := range peers {
			if p != thisPeer {
				out = append(out, p)
			}
		}

		emptyString, _ := bencode.EncodeBytes("")

		s.store[req.InfoHash] = out
		s.sendResponse(announceResponse{
			Interval: announceInterval,
			Peers:    emptyString,
			Peers6:   emptyString,
		}, w)

		return
	}

	compactPeers := []byte{}
	compactPeers6 := []byte{}

	for _, peer := range peers {
		if peer == thisPeer {
			continue
		}

		marshaled, err := peer.MarshalBinary()
		if err != nil {
			panic(err)
		}

		if peer.Ip.Is4() {
			compactPeers = append(compactPeers, marshaled...)
		} else {
			compactPeers6 = append(compactPeers6, marshaled...)
		}
	}

	formattedPeers, err := bencode.EncodeBytes(compactPeers)
	if err != nil {
		panic(err)
	}

	formattedPeers6, err := bencode.EncodeBytes(compactPeers6)
	if err != nil {
		panic(err)
	}

	s.sendResponse(announceResponse{
		Interval: announceInterval,
		Peers:    formattedPeers,
		Peers6:   formattedPeers6,
	}, w)
}

func parseRequest(values url.Values) (announceRequest, error) {
	req := announceRequest{}

	ih := values.Get("info_hash")
	pid := values.Get("peer_id")

	if len(ih) != 20 {
		return req, errors.New("invalid info_hash")
	}

	if len(pid) != 20 {
		return req, errors.New("invalid peer_id")
	}

	copy(req.InfoHash[:], ih)
	copy(req.PeerID[:], pid)

	port, err := strconv.ParseUint(values.Get("port"), 10, 16)
	if err != nil {
		return req, errors.New("invalid port")
	}
	if port == 0 {
		return req, errors.New("invalid port")
	}
	req.Port = uint16(port)

	switch req.Event {
	case "", "started", "stopped", "completed":
	default:
		return req, errors.New("invalid event")
	}

	req.Event = values.Get("event")
	req.Compact = values.Get("compact") == "1"

	if values.Get("numwant") != "" {
		numwant, err := strconv.ParseInt(values.Get("numwant"), 10, 64)
		if err != nil {
			return req, errors.New("invalid numwant")
		}
		req.Numwant = numwant
	} else {
		req.Numwant = -1
	}

	ip := values.Get("ip")
	if ip != "" {
		parsedIp, err := netip.ParseAddr(ip)
		if err != nil {
			return req, errors.New("invalid ip")
		}
		req.Ip = parsedIp
	} else {
		req.Ip = netip.Addr{}
	}

	return req, nil
}

func (s *HTTPServer) sendFailure(failure string, w http.ResponseWriter) {
	resp := announceResponse{}
	resp.Failure = failure
	s.sendResponse(resp, w)
}

func (s *HTTPServer) sendResponse(resp announceResponse, w http.ResponseWriter) {
	encodedResp, err := bencode.EncodeBytes(resp)
	if err != nil {
		panic(err)
	}
	w.Write(encodedResp)
}
