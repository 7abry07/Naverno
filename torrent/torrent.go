package torrent

import "GoBit/internal/metadata"

type Torrent struct {
	meta *metadata.Metadata

	closeC chan struct{}
	doneC  chan struct{}
}
