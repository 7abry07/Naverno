package torrent

import (
	"Naverno/internal/metadata"
	"log/slog"
)

// --------------- Structs -------------------

type Torrent struct {
	meta *metadata.Metadata

	logger *slog.Logger

	closeC chan struct{}
	doneC  chan struct{}
}
