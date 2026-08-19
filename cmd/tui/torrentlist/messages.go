package torrentlist

import (
	"Naverno/torrent"
	"context"
)

type AddTorrentMsg struct {
	Torrent *torrent.Torrent
}

type RemoveTorrentMsg struct {
	Torrent *torrent.Torrent
}

type StatsMsg struct {
	stats map[*torrent.Torrent]*torrent.TorrentStats
}

type TorrentEventMsg struct {
	Torrent *torrent.Torrent
	Context context.Context
	Event   any
}
