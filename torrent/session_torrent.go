package torrent

import (
	"fmt"
)

func (s *Session) AddTorrent(options TorrentOptions) (*Torrent, error) {
	if options.metadata == nil {
		return nil, fmt.Errorf("missing required metadata")
	}

	t, err := fromMetadata(s, options.metadata, options.SavePath, options.PieceSelectionStrategy)
	if err != nil {
		return nil, err
	}

	s.newTorrent <- t
	return t, nil
}

func (s *Session) RemoveTorrent(t *Torrent) {
	s.removeTorrent <- t
}

func (s *Session) handleNewTorrent(t *Torrent) {
	s.torrentsMut.Lock()
	defer s.torrentsMut.Unlock()

	s.torrents[t.infohash] = t
	go t.run()
}

func (s *Session) handleRemoveTorrent(t *Torrent) {
	t.Stop()
	s.torrentsMut.Lock()
	defer s.torrentsMut.Unlock()
	delete(s.torrents, t.infohash)
}

func (s *Session) stopTorrents() {
	for _, torr := range s.torrents {
		torr.Stop()
	}
}
