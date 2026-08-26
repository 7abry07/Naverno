package torrent

import (
	"Naverno/internal/magnet"
	"Naverno/internal/metadata"
	"bufio"
	"fmt"
	"os"
	"strings"
)

type PieceSelectionStrategy uint8

const (
	DEFAULT_PIECE_SELECTION PieceSelectionStrategy = iota
	SEQUENTIAL_PIECE_SELECTION
	RAREST_FIRST_PIECE_SELECTION
)

type TorrentOptions struct {
	metadata               *metadata.Metainfo
	magnet                 *magnet.Magnet
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

func FromURI(URI string) (TorrentOptions, error) {
	magnet, err := magnet.New(bufio.NewReader(strings.NewReader(URI)))
	if err != nil {
		return TorrentOptions{}, err
	}

	return TorrentOptions{
		magnet: magnet,
	}, nil
}
