package torrent

import (
	"Naverno/internal/announcer"
	"Naverno/internal/handshaker"
	"Naverno/internal/metadata"
	"Naverno/internal/peer"
	"Naverno/internal/tracker"
	"Naverno/internal/util"
	"fmt"
	"log/slog"
	"net"
	"net/netip"
	"time"
)

type Torrent struct {
	id         uint32
	pid        [20]byte
	port       uint16
	extensions [8]byte

	logger             *slog.Logger
	session            *Session
	meta               *metadata.Metadata
	announcer          *announcer.Announcer
	peers              []*peer.Peer
	outgoingHandshakes []*handshaker.OutgoingHandshaker

	downloaded int64
	uploaded   int64
	left       int64

	newConns                 chan net.Conn
	disconnectedPeers        chan *peer.Peer
	peerMessages             chan peer.PeerMessage
	torrentAnnounce          chan announcer.Torrent
	peersC                   chan []netip.AddrPort
	incomingHandshakeResults chan *handshaker.IncomingHandshaker
	outgoingHandshakeResults chan *handshaker.OutgoingHandshaker

	closeC chan struct{}
	doneC  chan struct{}
}

func newTorrentFromMetadata(sess *Session, id uint32, meta *metadata.Metadata) (*Torrent, error) {
	t := Torrent{
		session:                  sess,
		meta:                     meta,
		logger:                   sess.logger.With("TorrentID", id),
		peers:                    []*peer.Peer{},
		port:                     sess.port,
		downloaded:               0,
		uploaded:                 0,
		left:                     meta.Length,
		outgoingHandshakes:       []*handshaker.OutgoingHandshaker{},
		newConns:                 make(chan net.Conn),
		peerMessages:             make(chan peer.PeerMessage),
		disconnectedPeers:        make(chan *peer.Peer),
		torrentAnnounce:          make(chan announcer.Torrent),
		peersC:                   make(chan []netip.AddrPort),
		outgoingHandshakeResults: make(chan *handshaker.OutgoingHandshaker),
		incomingHandshakeResults: make(chan *handshaker.IncomingHandshaker),
		closeC:                   make(chan struct{}),
		doneC:                    make(chan struct{}),
		pid:                      sess.pid,
		id:                       id,
		extensions:               sess.extensions,
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

	for {
		select {
		case <-t.closeC:
			{
				t.ClosePeers()
				t.CloseHandshakes()
				t.announcer.Close()
				t.logger.Info("torrent -> stopped")
				return
			}
		case peers := <-t.peersC:
			{
				t.Dial(peers)
			}
		case <-t.torrentAnnounce:
			{
				t.torrentAnnounce <- announcer.Torrent{
					InfoHash:   t.meta.Infohash,
					PeerID:     t.pid,
					Downloaded: t.downloaded,
					Uploaded:   t.uploaded,
					Left:       t.left,
				}
			}
		case conn := <-t.newConns:
			{
				hs := handshaker.NewOutgoingHandshaker(conn)
				t.outgoingHandshakes = append(t.outgoingHandshakes, hs)
				go hs.Run(t.outgoingHandshakeResults, t.pid, t.meta.Infohash, t.extensions, time.Second*2)
				t.logger.Info("torrent -> started handshaker for connection", "Address", conn.RemoteAddr().String())
			}
		case res := <-t.outgoingHandshakeResults:
			{
				t.outgoingHandshakes = util.Remove(t.outgoingHandshakes, res, func(e1, e2 *handshaker.OutgoingHandshaker) bool { return e1 == e2 })
				if res.Error != nil {
					t.logger.Warn("torrent -> error during handshake", "Address", res.Conn.RemoteAddr().String(), "Error", res.Error.Error())
					continue
				}
				t.logger.Info("torrent -> connected to peer", "Peer", string(res.PeerID[:]), "Peer Count", len(t.peers))
				pe := peer.New(t.logger, res.Conn, res.PeerID, res.Extensions)
				t.peers = append(t.peers, pe)
				go pe.Run(t.peerMessages, t.disconnectedPeers)
			}
		case res := <-t.incomingHandshakeResults:
			{
				if res.Error != nil {
					t.logger.Warn("torrent -> error during handshake", "Address", res.Conn.RemoteAddr().String(), "Error", res.Error.Error())
					continue
				}
				t.logger.Info("torrent -> peer connected to us", "Peer", string(res.PeerID[:]), "Peer Count", len(t.peers))
				pe := peer.New(t.logger, res.Conn, res.PeerID, res.Extensions)
				t.peers = append(t.peers, pe)
				go pe.Run(t.peerMessages, t.disconnectedPeers)
			}
		case pe := <-t.disconnectedPeers:
			{
				t.peers = util.Remove(t.peers, pe, func(e1, e2 *peer.Peer) bool { return e1 == e2 })
				t.logger.Info("torrent -> peer disconnected", "Peer", string(pe.ID[:]), "Peer Count", len(t.peers))
				pe.Stop()
			}
		case pe := <-t.peerMessages:
			{
				t.logger.Info("torrent -> received message from peer", "PeerID", string(pe.ID[:]), "Message", pe.Message.ID().String())
			}
		}
	}
}

func (t *Torrent) Dial(addrs []netip.AddrPort) {
	for _, a := range addrs {
		go func() {
			conn, err := net.DialTimeout("tcp", a.String(), time.Second*5)
			if err != nil {
				t.logger.Warn("torrent -> error in connecting to remote peer", "Address", a.Addr().String(), "Error", err.Error())
				return
			}
			t.newConns <- conn
		}()
	}
}

func (t *Torrent) ClosePeers() {
	for _, p := range t.peers {
		p.Stop()
	}
}

func (t *Torrent) CloseHandshakes() {
	for _, hs := range t.outgoingHandshakes {
		hs.Close()
	}
}

func (t *Torrent) Stop() {
	close(t.closeC)
	<-t.doneC
}
