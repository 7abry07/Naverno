package udptracker

import (
	"Naverno/internal/tracker"
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"math/rand/v2"
	"net"
	"net/netip"
	"net/url"
	"sync"
	"time"
)

type UDPTracker struct {
	announce      url.URL
	pending       sync.Map
	connectionID  uint64
	lastConnected time.Time
}

func New(announceUrl url.URL) *UDPTracker {
	t := UDPTracker{
		announce: announceUrl,
		pending:  sync.Map{},
	}
	return &t
}

func (t *UDPTracker) Announce(ctx context.Context, req tracker.AnnounceRequest) (*tracker.AnnounceResponse, error) {
	addr, err := net.ResolveUDPAddr("udp", t.announce.Host)
	if err != nil {
		return nil, err
	}
	dialer := net.Dialer{}
	conn, err := dialer.DialUDP(ctx, "udp", netip.AddrPort{}, addr.AddrPort())
	if err != nil {
		return nil, err
	}

	if t.lastConnected.Equal(time.Time{}) || time.Since(t.lastConnected) > time.Minute {
		err := t.connect(ctx, conn)
		if err != nil {
			return nil, err
		}
	}
	r := announceRequest{
		connectionID: t.connectionID,
		infohash:     req.Infohash,
		peerID:       req.PeerID,
		downloaded:   uint64(req.Downloaded),
		uploaded:     uint64(req.Uploaded),
		left:         uint64(req.Left),
		event:        uint32(req.Event),
		ip:           binary.BigEndian.Uint32(req.Ip.AsSlice()),
		key:          0,
		numwant:      req.Numwant,
		port:         req.Port,
	}

	res, err := t.sendAnnounce(ctx, conn, r)
	if err != nil {
		return nil, err
	}

	re := tracker.AnnounceResponse{
		Interval: time.Duration(res.interval),
		Leechers: int64(res.leechers),
		Seeders:  int64(res.seeders),
		Peers:    res.peers,
	}

	return &re, nil
}

func (t *UDPTracker) URL() string {
	return t.announce.String()
}

func (t *UDPTracker) sendAnnounce(ctx context.Context, conn *net.UDPConn, req announceRequest) (announceResponse, error) {
	res := make(chan response)
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-done:
		case <-ctx.Done():
			conn.Close()
		}
	}()

	n := float64(0)
	timeout := time.NewTimer(time.Duration(0))
	<-timeout.C

	for n != 8 {
		tid := rand.Uint32()
		t.pending.Store(tid, res)
		req.transactionID = tid
		_, err := conn.Write(req.encode())
		if err != nil {
			return announceResponse{}, err
		}
		timeout.Reset(time.Second * time.Duration(15*math.Pow(2, n)))
		select {
		case <-ctx.Done():
			return announceResponse{}, ctx.Err()
		case r := <-res:
			if r.Action() != action_announce {
			}
			switch r.Action() {
			case action_announce:
				return r.(announceResponse), nil
			case action_error:
				return announceResponse{}, fmt.Errorf("%v", r.(errorResponse).message)
			default:
				panic("got non announce response from announce request in UDP tracker")
			}

		case <-timeout.C:
		}
		n++
		t.pending.Delete(tid)
	}

	return announceResponse{}, fmt.Errorf("announce timeout")
}

func (t *UDPTracker) connect(ctx context.Context, conn *net.UDPConn) error {
	res := make(chan response)
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-done:
		case <-ctx.Done():
			conn.Close()
		}
	}()

	n := float64(0)
	timeout := time.NewTimer(time.Duration(0))
	<-timeout.C

	for n != 8 {
		tid := rand.Uint32()
		t.pending.Store(tid, res)
		req := connectRequest{tid}
		_, err := conn.Write(req.encode())
		if err != nil {
			return err
		}
		timeout.Reset(time.Second * time.Duration(15*math.Pow(2, n)))
		select {
		case <-ctx.Done():
			return ctx.Err()
		case r := <-res:
			switch r.Action() {
			case action_connect:
				t.connectionID = r.(connectResponse).connectionID
				t.lastConnected = time.Now()
				return nil
			case action_error:
				return fmt.Errorf("%v", r.(errorResponse).message)
			default:
				panic("got non connect response from connect request in UDP tracker")
			}
		case <-timeout.C:
		}
		n++
		t.pending.Delete(tid)
	}

	return fmt.Errorf("connect timeout")
}
