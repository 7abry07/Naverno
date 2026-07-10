package testserver

import (
	"Naverno/internal/tracker"
	"errors"
	"log/slog"
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

type peer struct {
	Ip   string `bencode:"ip"`
	Port uint16 `bencode:"port"`
}

type HTTPServer struct {
	store  map[[20]byte][]peer
	m      sync.RWMutex
	logger *slog.Logger
}

func StartHttp(logger *slog.Logger) {
	s := HTTPServer{}
	s.store = make(map[[20]byte][]peer)
	s.m = sync.RWMutex{}
	s.logger = logger

	listener, err := net.Listen("tcp", ":8000")
	if err != nil {
		s.logger.Error("error in opening listening socket", "error", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/announce", s.announce)
	go http.Serve(listener, mux)
}

func (s *HTTPServer) announce(w http.ResponseWriter, r *http.Request) {
	s.m.Lock()
	defer s.m.Unlock()
	req, err := parseRequest(r.URL.Query())
	if err != nil {
		s.sendFailure(err.Error(), w)
	}

	thisPeer := peer{}
	thisPeer.Port = req.Port
	if (req.Ip != netip.Addr{}) {
		thisPeer.Ip = req.Ip.String()
	} else {
		thisPeer.Ip = r.RemoteAddr
	}

	peers, ok := s.store[req.InfoHash]
	if !ok {
		s.store[req.InfoHash] = []peer{}
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
		out := []peer{}

		for _, p := range peers {
			if p != thisPeer {
				out = append(out, p)
			}
		}

		s.store[req.InfoHash] = out
		s.sendResponse(announceResponse{
			Interval: announceInterval,
			Peers:    bencode.RawMessage{'0', ':'},
			Peers6:   bencode.RawMessage{'0', ':'},
		}, w)

		return
	}

	formattedPeers := bencode.RawMessage{}
	formattedPeers6 := bencode.RawMessage{}

	compactPeers := []byte{}
	compactPeers6 := []byte{}

	for _, peer := range peers {
		if peer == thisPeer {
			continue
		}

		compactPeer, err := tracker.NewCompactPeer(peer.Ip, peer.Port)
		if err != nil {
			s.logger.Error("error in creating compact peer", "error", err)
			return
		}
		marshaled, err := compactPeer.MarshalBinary()
		if err != nil {
			s.logger.Error("error in marshaling compact peer", "error", err)
			return
		}

		if len(marshaled) == 18 {
			compactPeers6 = append(compactPeers6, marshaled...)
		} else if len(marshaled) == 6 {
			compactPeers = append(compactPeers, marshaled...)
		} else {
			s.logger.Error("error in marshaling compact peer", "error", "ip:port len is neither 6 or 18")
			return
		}
	}

	formattedPeers, err = bencode.EncodeBytes(compactPeers)
	if err != nil {
		s.logger.Error("error in marshaling compact peers", "error", err)
		return
	}

	formattedPeers6, err = bencode.EncodeBytes(compactPeers6)
	if err != nil {
		s.logger.Error("error in marshaling compact peers", "error", err)
		return
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

	copy(req.InfoHash[:], ih)
	copy(req.PeerID[:], pid)

	port, err := strconv.ParseUint(values.Get("port"), 10, 16)
	if err != nil {
		return req, errors.New("invalid port")
	}
	req.Port = uint16(port)

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

	if err := ValidateAnnounceRequest(req); err != nil {
		return req, err
	}

	return req, nil
}

func ValidateAnnounceRequest(req announceRequest) error {
	if len(req.InfoHash) != 20 {
		return errors.New("invalid info_hash")
	}

	if len(req.PeerID) != 20 {
		return errors.New("invalid peer_id")
	}

	if req.Port == 0 {
		return errors.New("invalid port")
	}

	switch req.Event {
	case "", "started", "stopped", "completed":
	default:
		return errors.New("invalid event")
	}

	return nil
}

func (s *HTTPServer) sendFailure(failure string, w http.ResponseWriter) {
	resp := announceResponse{}
	resp.Failure = failure
	s.sendResponse(resp, w)
}

func (s *HTTPServer) sendResponse(resp announceResponse, w http.ResponseWriter) {
	encodedResp, err := bencode.EncodeBytes(resp)
	if err != nil {
		s.logger.Error("error in encoding response", "error", err)
		return
	}
	w.Write(encodedResp)
}
