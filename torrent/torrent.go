package torrent

import (
	"Naverno/internal/announcer"
	"Naverno/internal/bitfield"
	"Naverno/internal/choker"
	"Naverno/internal/handshaker"
	"Naverno/internal/metadata"
	"Naverno/internal/peer"
	"Naverno/internal/peerprotocol"
	"Naverno/internal/picker"
	"Naverno/internal/piece"
	"Naverno/internal/piecedownloader"
	"Naverno/internal/storage"
	"Naverno/internal/storage/posixstorage"
	"Naverno/internal/tracker"
	"fmt"
	"log/slog"
	"net"
	"net/netip"
	"path/filepath"
	"sync"
	"time"
)

type Torrent struct {
	id         []byte
	downloaded int64
	uploaded   int64
	left       int64
	savePath   string

	downloadedSince uint64
	uploadedSince   uint64
	uploadRate      uint64
	downloadRate    uint64
	rateMut         sync.Mutex

	session            *Session
	storage            storage.Storage
	picker             *picker.Picker
	choker             *choker.Choker
	bitset             *bitfield.Bitfield
	logger             *slog.Logger
	meta               *metadata.Metadata
	announcer          *announcer.Announcer
	pieces             []*piece.Piece
	peers              map[[20]byte]*peer.Peer
	outgoing           map[*handshaker.OutgoingHandshaker]struct{}
	pendingRequests    map[peerprotocol.Request][20]byte
	downloaders        map[*peer.Peer]*piecedownloader.PieceDownloader
	stalledDownloaders map[*piece.Piece]*piecedownloader.PieceDownloader
	newConns           chan net.Conn
	disconnectedPeers  chan *peer.Peer
	peerMessages       chan peer.PeerMessage
	torrentAnnounce    chan announcer.Torrent
	writeResults       chan storage.WriteResult
	hashResults        chan storage.HashResult
	readResults        chan storage.ReadResult
	incomingResults    chan *handshaker.IncomingHandshaker
	outgoingResults    chan *handshaker.OutgoingHandshaker
	peersC             chan []netip.AddrPort
	statsRequest       chan TorrentStats
	chokerEvents       chan any

	err    error
	closeC chan struct{}
	doneC  chan struct{}
}

func newTorrentFromMetadata(sess *Session, meta *metadata.Metadata, savePath string) (*Torrent, error) {
	if len(meta.Files) > 1 {
		savePath = filepath.Join(savePath, meta.Name)
	}
	t := Torrent{
		session:            sess,
		meta:               meta,
		logger:             sess.logger.With("TorrentID", fmt.Sprintf("%X", meta.Infohash[:4])),
		savePath:           savePath,
		peers:              make(map[[20]byte]*peer.Peer),
		outgoing:           make(map[*handshaker.OutgoingHandshaker]struct{}),
		downloaders:        make(map[*peer.Peer]*piecedownloader.PieceDownloader),
		stalledDownloaders: make(map[*piece.Piece]*piecedownloader.PieceDownloader),
		pendingRequests:    make(map[peerprotocol.Request][20]byte),
		downloaded:         0,
		uploaded:           0,
		left:               meta.Length,
		picker:             picker.New(uint32(meta.PieceCount)),
		choker:             choker.New(time.Second*10, time.Second*30),
		pieces:             piece.NewPieces(meta),
		bitset:             bitfield.New(uint32(meta.PieceCount)),
		storage:            posixstorage.New(sess.logger, meta.Files, savePath),
		writeResults:       make(chan storage.WriteResult),
		hashResults:        make(chan storage.HashResult),
		readResults:        make(chan storage.ReadResult),
		newConns:           make(chan net.Conn),
		peerMessages:       make(chan peer.PeerMessage),
		disconnectedPeers:  make(chan *peer.Peer),
		torrentAnnounce:    make(chan announcer.Torrent),
		peersC:             make(chan []netip.AddrPort),
		chokerEvents:       make(chan any),
		statsRequest:       make(chan TorrentStats),
		outgoingResults:    make(chan *handshaker.OutgoingHandshaker),
		incomingResults:    make(chan *handshaker.IncomingHandshaker),
		closeC:             make(chan struct{}),
		doneC:              make(chan struct{}),
		id:                 meta.Infohash[:4],
	}

	trackers := [][]tracker.Tracker{}
	for _, urls := range meta.AnnounceList {
		tier := []tracker.Tracker{}
		for _, url := range urls {
			tr, err := sess.trackerManager.Get(url.String())
			if err != nil {
				t.logger.Warn("torrent -> couldn't get tracker", "Tracker URL", url.String(), "Error", err.Error())
				continue
			}
			tier = append(tier, tr)
		}
		trackers = append(trackers, tier)
	}

	t.announcer = announcer.New(t.logger, trackers, sess.port)

	return &t, nil
}

func (t *Torrent) run() {
	defer close(t.doneC)

	go t.announcer.Run(t.torrentAnnounce, t.peersC)
	go t.choker.Run(t.chokerEvents)

	rateTick := time.NewTicker(time.Second)

	for {
		select {
		case <-t.closeC:
			t.closePeers()
			t.closeHandshakes()
			t.closeAnnouncer()
			t.storage.StopPendingJobs()
			t.choker.Close()
			return
		case <-rateTick.C:
			t.rateMut.Lock()
			t.uploadRate = t.uploadedSince * 8
			t.downloadRate = t.downloadedSince * 8
			t.downloadedSince = 0
			t.uploadedSince = 0
			t.rateMut.Unlock()
		case req := <-t.statsRequest:
			t.fillStats(&req)
			t.statsRequest <- req
		case ev := <-t.chokerEvents:
			t.handleChokerEvent(ev)
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
		case res := <-t.writeResults:
			t.handleWriteResult(res)
		case res := <-t.hashResults:
			t.handleHashResult(res)
		case res := <-t.readResults:
			t.handleReadResult(res)
		case p := <-t.peerMessages:
			t.handlePeerMessage(p)
		}
	}
}

func (t *Torrent) Metadata() (metadata.Metadata, bool) {
	if t.meta != nil {
		return *t.meta, true
	}
	return metadata.Metadata{}, false
}
