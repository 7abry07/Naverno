package tracker

import (
	"net/netip"
	"net/url"
)

// --------------- Constants -------------------

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

// --------------- Structs -------------------

type AnnounceRequest struct {
	Infohash   [20]byte
	PeerID     [20]byte
	Downloaded uint64
	Uploaded   uint64
	Left       uint64
	Ip         netip.Addr
	Port       uint16
	Key        uint32
	Event      trackerEvent
}

type AnnounceResponse struct {
	Failure     *string
	Warning     *string
	MinInterval uint32
	Interval    uint32
	Complete    int64
	Incomplete  int64
	Downloaded  int64
	PeerList    []peer
}

type peer struct {
	Ip   netip.Addr
	Port uint16
}

// -------------- Interfaces ------------------

type Tracker interface {
	Announce(r AnnounceRequest) (AnnounceResponse, error)
}

// -------------- Functions ------------------

func NewTracker(announceUrl string) (Tracker, error) {
	parsedUrl, err := url.Parse(announceUrl)
	if err != nil {
		return nil, err
	}

	switch parsedUrl.Scheme {
	case "https":
		fallthrough
	case "http":
		httptracker := httpTracker{}
		httptracker.announce = parsedUrl
		return httptracker, nil
	}
	return nil, InvalidSchemeErr
}
