package announcer

type Torrent struct {
	InfoHash   [20]byte
	PeerID     [20]byte
	Downloaded int64
	Uploaded   int64
	Left       int64
}
