package tracker

import (
	"context"
	"net/netip"
	"time"
)

type Tracker interface {
	Announce(ctx context.Context, r AnnounceRequest) (*AnnounceResponse, error)
	URL() string
}

type AnnounceRequest struct {
	Infohash   [20]byte
	PeerID     [20]byte
	Downloaded int64
	Uploaded   int64
	Left       int64
	Ip         netip.Addr
	Port       uint16
	Numwant    uint32
	Event      TrackerEvent
}

type AnnounceResponse struct {
	MinInterval    time.Duration
	Interval       time.Duration
	Leechers       int64
	Seeders        int64
	WarningMessage string
	Peers          []netip.AddrPort
}

type Error struct {
	Reason  string
	RetryIn time.Duration
}

func (e Error) Error() string {
	return e.Reason
}
