package torrent

import (
	"Naverno/internal/handshaker"
	"Naverno/internal/metadata"
	"Naverno/internal/peer"
	"Naverno/internal/tracker"
	"Naverno/internal/util"
	"context"
	"fmt"
	"log/slog"
	"net"
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
	trackers           []tracker.Tracker
	peers              []*peer.Peer
	outgoingHandshakes []*handshaker.OutgoingHandshaker

	downloaded int64
	uploaded   int64
	left       int64

	newConns                 chan net.Conn
	disconnectedPeers        chan *peer.Peer
	peerMessages             chan peer.PeerMessage
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
		newConns:                 make(chan net.Conn),
		peerMessages:             make(chan peer.PeerMessage),
		disconnectedPeers:        make(chan *peer.Peer),
		outgoingHandshakes:       []*handshaker.OutgoingHandshaker{},
		outgoingHandshakeResults: make(chan *handshaker.OutgoingHandshaker),
		incomingHandshakeResults: make(chan *handshaker.IncomingHandshaker),
		closeC:                   make(chan struct{}),
		doneC:                    make(chan struct{}),
		pid:                      sess.pid,
		id:                       id,
		extensions:               sess.extensions,
	}

	for _, urls := range meta.AnnounceList {
		for _, url := range urls {
			tr, err := sess.trackerManager.Get(url.String())
			if err != nil {
				return nil, fmt.Errorf("error in getting tracker implementation -> %v", err)
			}
			t.trackers = append(t.trackers, tr)
		}
	}

	return &t, nil
}

func (t *Torrent) run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)

	defer close(t.doneC)
	defer cancel()

	announceReq := tracker.AnnounceRequest{
		Infohash:   t.meta.Infohash,
		PeerID:     t.pid,
		Downloaded: t.downloaded,
		Uploaded:   t.uploaded,
		Left:       t.left,
		Event:      tracker.TRACKER_STARTED,
		Port:       t.port,
	}

	for _, tr := range t.trackers {
		res, err := tr.Announce(ctx, announceReq)
		if err != nil {
			t.logger.Warn("torrent -> error in announcing to tracker", "Tracker URL", tr.URL(), "Error", err.Error())
			continue
		}
		for _, p := range res.Peers {
			go func() {
				conn, err := net.DialTimeout("tcp", p.String(), time.Second*5)
				if err != nil {
					t.logger.Warn("torrent -> error in connecting to peer", "Address", p.Addr().String(), "Error", err.Error())
					return
				}
				t.newConns <- conn
			}()
		}
	}

	for {
		select {
		case <-t.closeC:
			for _, p := range t.peers {
				p.Stop()
				t.logger.Info("torrent -> active peer closed", "Peer", string(p.ID[:]))
			}
			for _, hs := range t.outgoingHandshakes {
				hs.Close()
				t.logger.Info("torrent -> running handshake closed", "Address", hs.Conn.RemoteAddr())
			}

			announceStop := tracker.AnnounceRequest{
				Infohash:   t.meta.Infohash,
				PeerID:     t.pid,
				Downloaded: t.downloaded,
				Uploaded:   t.uploaded,
				Left:       t.left,
				Event:      tracker.TRACKER_STOPPED,
				Port:       t.port,
				Numwant:    0,
			}
			for _, tr := range t.trackers {
				tr.Announce(ctx, announceStop)
			}
			t.logger.Info("torrent -> stopped")
			return
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
				t.logger.Info("torrent -> received message from peer", "Peer", string(pe.ID[:]), "Message", pe.Message.ID().String())
			}
		}
	}
}

func (t *Torrent) Stop() {
	close(t.closeC)
	<-t.doneC
}
