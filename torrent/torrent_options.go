package torrent

import (
	"Naverno/internal/metadata"
	"fmt"
	"os"
)

type PieceSelectionStrategy uint8

const (
	DEFAULT_PIECE_SELECTION PieceSelectionStrategy = iota
	SEQUENTIAL_PIECE_SELECTION
	RAREST_FIRST_PIECE_SELECTION
)

type TorrentOptions struct {
	metadata               *metadata.Metainfo
	SavePath               string
	PieceSelectionStrategy PieceSelectionStrategy
}

func FromFile(file string) (TorrentOptions, error) {
	f, err := os.Open(file)
	if err != nil {
		return TorrentOptions{}, fmt.Errorf("error opening torrent file -> %v", err)
	}

	meta, err := metadata.New(f)
	if err != nil {
		return TorrentOptions{}, fmt.Errorf("error creating torrent metadata -> %v", err)
	}

	return TorrentOptions{
		metadata: meta,
	}, nil
}
