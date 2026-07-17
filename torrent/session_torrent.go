package torrent

import (
	"Naverno/internal/metadata"
	"fmt"
	"os"
)

func (s *Session) handleNewTorrent(t *Torrent) {
	s.torrentsMut.Lock()
	defer s.torrentsMut.Unlock()

	s.torrents[t.meta.Infohash] = t
	go t.run()
}

func (s *Session) handleRemoveTorrent(t *Torrent) {
	t.Stop()
	s.torrentsMut.Lock()
	defer s.torrentsMut.Unlock()
	delete(s.torrents, t.meta.Infohash)
}

func (s *Session) NewTorrentFromFile(path string) (*Torrent, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening torrent file -> %v", err)
	}

	meta, err := metadata.New(file)
	if err != nil {
		return nil, fmt.Errorf("error creating torrent metadata -> %v", err)
	}

	t, err := newTorrentFromMetadata(s, s.currentTid, meta)
	if err != nil {
		return nil, err
	}

	s.currentTid++
	s.newTorrent <- t

	return t, nil
}

func (s *Session) stopTorrents() {
	for _, torr := range s.torrents {
		torr.Stop()
	}
}
