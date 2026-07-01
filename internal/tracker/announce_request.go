package tracker

import (
	"net/netip"
)

// --------------- Structs -------------------

type AnnounceRequest struct {
	Infohash   [20]byte
	PeerID     [20]byte
	Downloaded uint64
	Uploaded   uint64
	Left       uint64
	Numwant    uint32
	Ip         netip.Addr
	Port       uint16
	Key        uint32
	NoPID      uint8
	Compact    uint8
	Event      trackerEvent
}
