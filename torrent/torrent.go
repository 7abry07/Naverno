package torrent

import (
	"Naverno/internal/metadata"
	"Naverno/internal/tracker"
	"log/slog"
)

// --------------- Structs -------------------

type Torrent struct {
	pid [20]byte

	session  *Session
	meta     *metadata.Metadata
	trackers [][]tracker.Tracker

	logger *slog.Logger

	closeC chan struct{}
	doneC  chan struct{}
}
