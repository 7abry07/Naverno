package torrent

import (
	"Naverno/internal/announcer"
	"Naverno/internal/bitfield"
	"Naverno/internal/handshaker"
	"Naverno/internal/metadata"
	"Naverno/internal/peer"
	"Naverno/internal/picker"
	"Naverno/internal/picker/sequentialpicker"
	"Naverno/internal/piece"
	"Naverno/internal/piecedownloader"
	"Naverno/internal/piecewriter"
	"Naverno/internal/storage"
	"Naverno/internal/storage/filestorage"
	"Naverno/internal/tracker"
	"log/slog"
	"net"
	"net/netip"
)

type Torrent struct {
	id         uint32
	pid        [20]byte
	port       uint16
	extensions [8]byte

	session            *Session
	storage            storage.Storage
	picker             picker.Picker
	pieces             []*piece.Piece
	bitset             *bitfield.Bitfield
	logger             *slog.Logger
	meta               *metadata.Metadata
	announcer          *announcer.Announcer
	peers              map[*peer.Peer]struct{}
	outgoing           map[*handshaker.OutgoingHandshaker]struct{}
	writers            map[*piece.Piece]*piecewriter.PieceWriter
	downloaders        map[*peer.Peer]*piecedownloader.PieceDownloader
	stalledDownloaders map[*piece.Piece]*piecedownloader.PieceDownloader

	downloaded int64
	uploaded   int64
	left       int64

	newConns          chan net.Conn
	disconnectedPeers chan *peer.Peer
	peerMessages      chan peer.PeerMessage
	torrentAnnounce   chan announcer.Torrent
	writerResults     chan *piecewriter.PieceWriter
	incomingResults   chan *handshaker.IncomingHandshaker
	outgoingResults   chan *handshaker.OutgoingHandshaker
	peersC            chan []netip.AddrPort

	closeC chan struct{}
	doneC  chan struct{}
}

func newTorrentFromMetadata(sess *Session, id uint32, meta *metadata.Metadata) (*Torrent, error) {
	pieces := piece.NewPieces(meta)
	t := Torrent{
		session:            sess,
		meta:               meta,
		logger:             sess.logger.With("TorrentID", id),
		storage:            filestorage.New(meta.Files, sess.path),
		peers:              make(map[*peer.Peer]struct{}),
		outgoing:           make(map[*handshaker.OutgoingHandshaker]struct{}),
		downloaders:        make(map[*peer.Peer]*piecedownloader.PieceDownloader),
		stalledDownloaders: make(map[*piece.Piece]*piecedownloader.PieceDownloader),
		writers:            make(map[*piece.Piece]*piecewriter.PieceWriter),
		port:               sess.port,
		downloaded:         0,
		uploaded:           0,
		left:               meta.Length,
		picker:             sequentialpicker.NewSequentialPicker(pieces),
		pieces:             pieces,
		bitset:             bitfield.New(uint32(meta.PieceCount)),
		writerResults:      make(chan *piecewriter.PieceWriter),
		newConns:           make(chan net.Conn),
		peerMessages:       make(chan peer.PeerMessage),
		disconnectedPeers:  make(chan *peer.Peer),
		torrentAnnounce:    make(chan announcer.Torrent),
		peersC:             make(chan []netip.AddrPort),
		outgoingResults:    make(chan *handshaker.OutgoingHandshaker),
		incomingResults:    make(chan *handshaker.IncomingHandshaker),
		closeC:             make(chan struct{}),
		doneC:              make(chan struct{}),
		pid:                sess.pid,
		id:                 id,
		extensions:         sess.extensions,
	}

	trackers := [][]tracker.Tracker{}
	for _, urls := range meta.AnnounceList {
		tier := []tracker.Tracker{}
		for _, url := range urls {
			tr, err := sess.trackerManager.Get(url.String())
			if err != nil {
				t.logger.Warn("torrent -> couldn't get tracker implementation", "Tracker URL", url.String(), "Error", err.Error())
				continue
			}
			tier = append(tier, tr)
		}
		trackers = append(trackers, tier)
	}

	t.announcer = announcer.New(t.logger, trackers, t.port)

	return &t, nil
}

func (t *Torrent) run() {
	defer close(t.doneC)

	go t.announcer.Run(t.torrentAnnounce, t.peersC)

	for {
		select {
		case <-t.closeC:
			t.closePeers()
			t.closeHandshakes()
			t.closeAnnouncer()
			t.closeWriters()
			return
		case conn := <-t.newConns:
			t.handleNewConn(conn)
		case p := <-t.disconnectedPeers:
			t.handleDisconnected(p)
		case peers := <-t.peersC:
			t.dial(peers)
		case <-t.torrentAnnounce:
			t.handleAnnounce()
		case res := <-t.outgoingResults:
			t.handleOutgoingResult(res)
		case res := <-t.incomingResults:
			t.handleIncomingResult(res)
		case res := <-t.writerResults:
			if res.Err != nil {
				t.logger.Info("torrent -> error in piece writer", "Error", res.Err)
			}
		case p := <-t.peerMessages:
			t.handlePeerMessage(p)
		}
	}
}

func (t *Torrent) Stop() {
	close(t.closeC)
	<-t.doneC
	t.logger.Info("torrent -> stopped")
}
