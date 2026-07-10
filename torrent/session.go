package torrent

import (
	"Naverno/internal/metadata"
	"Naverno/internal/tracker"
	"Naverno/internal/trackermanager"
	"fmt"
	"log/slog"
	"os"
)

// --------------- Structs -------------------

type Session struct {
	torrents map[[20]byte]*Torrent

	trackerManager *trackermanager.TrackerManager
	logger         *slog.Logger

	closeC chan struct{}
	doneC  chan struct{}
}

func NewSession(logger *slog.Logger) *Session {
	if logger == nil {
		panic("cannot pass nil logger to session")
	}

	s := Session{}

	s.torrents = make(map[[20]byte]*Torrent)

	s.trackerManager = trackermanager.New(logger)
	s.logger = logger
	s.closeC = make(chan struct{})
	s.doneC = make(chan struct{})

	return &s
}

func (s *Session) NewTorrentFromFile(path string) (*Torrent, error) {
	t := Torrent{}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening torrent file -> %v", err)
	}

	meta, err := metadata.New(file)
	if err != nil {
		return nil, fmt.Errorf("error creating torrent metadata -> %v", err)
	}

	t.session = s
	t.meta = meta
	t.closeC = make(chan struct{})
	t.doneC = make(chan struct{})

	pid := "-NV0001-djgncuteodpq"
	copy(t.pid[:], []byte(pid))

	for _, urls := range meta.AnnounceList {
		trs := []tracker.Tracker{}
		for _, url := range urls {
			tr, err := s.trackerManager.Get(url.String())
			if err != nil {
				return nil, fmt.Errorf("error in getting tracker implementation -> %v", err)
			}
			trs = append(trs, tr)
		}
		t.trackers = append(t.trackers, trs)
	}

	s.torrents[t.meta.Infohash] = &t

	return &t, nil
}
