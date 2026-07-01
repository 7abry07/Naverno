package torrent

import (
	"GoBit/internal/metadata"
	"log/slog"
)

// --------------- Structs -------------------

type Torrent struct {
	meta *metadata.Metadata

	logger *slog.Logger

	closeC chan struct{}
	doneC  chan struct{}
}
