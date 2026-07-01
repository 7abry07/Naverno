package torrent

type Session struct {
	torrents map[[20]byte]Torrent

	closeC chan struct{}
	doneC  chan struct{}
}
