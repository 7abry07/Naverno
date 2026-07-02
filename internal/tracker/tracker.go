package tracker

import (
	"context"
	"net/netip"
	"time"
)

type trackerEvent uint8

const (
	TRACKER_NONE trackerEvent = iota
	TRACKER_STARTED
	TRACKER_COMPLETED
	TRACKER_STOPPED
)

var trackerEventName = map[trackerEvent]string{
	TRACKER_NONE:      "",
	TRACKER_STARTED:   "started",
	TRACKER_COMPLETED: "completed",
	TRACKER_STOPPED:   "stopped",
}

func (e trackerEvent) String() string {
	return trackerEventName[e]
}

type Tracker interface {
	Announce(ctx context.Context, r AnnounceRequest) (*AnnounceResponse, error)
	URL() string
}

type AnnounceRequest struct {
	Infohash   [20]byte
	PeerID     [20]byte
	Downloaded uint64
	Uploaded   uint64
	Left       uint64
	Ip         netip.Addr
	Port       uint16
	Numwant    uint32
	Event      trackerEvent
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
