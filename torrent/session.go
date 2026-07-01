package torrent

import "log/slog"

// --------------- Structs -------------------

type Session struct {
	torrents map[[20]byte]Torrent

	logger *slog.Logger

	closeC chan struct{}
	doneC  chan struct{}
}
