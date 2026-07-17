package torrent

import (
	"Naverno/internal/announcer"
	"Naverno/internal/handshaker"
	"Naverno/internal/metadata"
	"Naverno/internal/peer"
	"Naverno/internal/tracker"
	"fmt"
	"log/slog"
	"net"
	"net/netip"
)

type Torrent struct {
	id         uint32
	pid        [20]byte
	port       uint16
	extensions [8]byte

	session   *Session
	logger    *slog.Logger
	meta      *metadata.Metadata
	announcer *announcer.Announcer
	outgoing  []*handshaker.OutgoingHandshaker
	peers     []*peer.Peer

	downloaded int64
	uploaded   int64
	left       int64

	newConns          chan net.Conn
	disconnectedPeers chan *peer.Peer
	peerMessages      chan peer.PeerMessage
	torrentAnnounce   chan announcer.Torrent
	peersC            chan []netip.AddrPort
	incomingResults   chan *handshaker.IncomingHandshaker
	outgoingResults   chan *handshaker.OutgoingHandshaker

	closeC chan struct{}
	doneC  chan struct{}
}

func newTorrentFromMetadata(sess *Session, id uint32, meta *metadata.Metadata) (*Torrent, error) {
	t := Torrent{
		session:           sess,
		meta:              meta,
		logger:            sess.logger.With("TorrentID", id),
		peers:             []*peer.Peer{},
		port:              sess.port,
		downloaded:        0,
		uploaded:          0,
		left:              meta.Length,
		outgoing:          []*handshaker.OutgoingHandshaker{},
		newConns:          make(chan net.Conn),
		peerMessages:      make(chan peer.PeerMessage),
		disconnectedPeers: make(chan *peer.Peer),
		torrentAnnounce:   make(chan announcer.Torrent),
		peersC:            make(chan []netip.AddrPort),
		outgoingResults:   make(chan *handshaker.OutgoingHandshaker),
		incomingResults:   make(chan *handshaker.IncomingHandshaker),
		closeC:            make(chan struct{}),
		doneC:             make(chan struct{}),
		pid:               sess.pid,
		id:                id,
		extensions:        sess.extensions,
	}

	trackers := []tracker.Tracker{}
	for _, urls := range meta.AnnounceList {
		for _, url := range urls {
			tr, err := sess.trackerManager.Get(url.String())
			if err != nil {
				return nil, fmt.Errorf("error in getting tracker implementation -> %v", err)
			}
			trackers = append(trackers, tr)
		}
	}

	t.announcer = announcer.NewAnnouncer(t.logger, t.torrentAnnounce, trackers, t.port)

	return &t, nil
}

func (t *Torrent) run() {
	go t.announcer.Run(t.peersC)

	defer close(t.doneC)
	defer t.closePeers()
	defer t.closeHandshakes()
	defer t.closeAnnouncer()
	defer t.logger.Info("torrent -> stopped")

	for {
		select {
		case <-t.closeC:
			return
		case conn := <-t.newConns:
			t.handleNewConn(conn)
		case p := <-t.disconnectedPeers:
			t.handleDisconnected(p)
		case peers := <-t.peersC:
			t.Dial(peers)
		case <-t.torrentAnnounce:
			t.handleAnnounce()
		case res := <-t.outgoingResults:
			t.handleOutgoingResult(res)
		case res := <-t.incomingResults:
			t.handleIncomingResult(res)
		case p := <-t.peerMessages:
			t.logger.Info("torrent -> received message from peer", "PeerID", string(p.ID[:]), "Message", p.Message.ID().String())
		}
	}
}

func (t *Torrent) Stop() {
	close(t.closeC)
	<-t.doneC
}
