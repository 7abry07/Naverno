package tracker

import "net/netip"

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

type peer struct {
	Ip   netip.Addr
	Port uint16
}

// -------------- Interfaces ------------------

type Tracker interface {
	Announce() announceResponse
}
