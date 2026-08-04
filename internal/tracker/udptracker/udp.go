package udptracker

import (
	"Naverno/internal/tracker"
	"context"
	"encoding/binary"
	"net/netip"
	"net/url"
	"time"
)

type UDPTracker struct {
	url           url.URL
	connectionID  uint64
	lastConnected time.Time
	transport     *UDPTransport
}

func New(announceUrl url.URL, transport *UDPTransport) *UDPTracker {
	t := UDPTracker{
		url:       announceUrl,
		transport: transport,
	}
	return &t
}

func (t *UDPTracker) Announce(ctx context.Context, req tracker.AnnounceRequest) (*tracker.AnnounceResponse, error) {
	if t.lastConnected.Equal(time.Time{}) || time.Since(t.lastConnected) > time.Minute {
		connectReq := NewUDPTransportRequest(ctx, t.url, connectRequest{})
		res, err := t.transport.Do(connectReq)
		if err != nil {
			return nil, err
		}
		connectRes := res.(connectResponse)
		t.connectionID = connectRes.connectionID
		t.lastConnected = time.Now()
	}

	ip := uint32(0)
	if (req.Ip != netip.Addr{}) {
		arr := req.Ip.As4()
		slice := make([]byte, 4)
		copy(slice, arr[:])
		ip = binary.BigEndian.Uint32(slice)
	}
	announceReq := NewUDPTransportRequest(ctx, t.url, announceRequest{
		connectionID: t.connectionID,
		infohash:     req.Infohash,
		peerID:       req.PeerID,
		downloaded:   uint64(req.Downloaded),
		uploaded:     uint64(req.Uploaded),
		left:         uint64(req.Left),
		event:        uint32(req.Event),
		key:          binary.BigEndian.Uint32(req.PeerID[16:20]),
		ip:           ip,
		numwant:      req.Numwant,
		port:         req.Port,
	})
	res, err := t.transport.Do(announceReq)
	if err != nil {
		return nil, err
	}
	announceRes := res.(announceResponse)

	return &tracker.AnnounceResponse{
		Interval: time.Second * time.Duration(announceRes.interval),
		Leechers: int64(announceRes.leechers),
		Seeders:  int64(announceRes.seeders),
		Peers:    announceRes.peers,
	}, nil
}

func (t *UDPTracker) URL() string {
	return t.url.String()
}
