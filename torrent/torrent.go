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
	pid [20]byte

	session            *Session
	meta               *metadata.Metadata
	trackers           []tracker.Tracker
	peers              []*peer.Peer
	outgoingHandshakes []*handshaker.OutgoingHandshaker

	newConns                 chan net.Conn
	disconnectedPeers        chan *peer.Peer
	peerMessages             chan peer.PeerMessage
	incomingHandshakeResults chan *handshaker.IncomingHandshaker
	outgoingHandshakeResults chan *handshaker.OutgoingHandshaker

	logger *slog.Logger

	closeC chan struct{}
	doneC  chan struct{}
}

func newTorrentFromMetadata(sess *Session, meta *metadata.Metadata) (*Torrent, error) {
	t := Torrent{}

	t.session = sess
	t.meta = meta
	t.logger = sess.logger
	t.peers = []*peer.Peer{}
	t.newConns = make(chan net.Conn)
	t.peerMessages = make(chan peer.PeerMessage)
	t.disconnectedPeers = make(chan *peer.Peer)
	t.outgoingHandshakes = []*handshaker.OutgoingHandshaker{}
	t.outgoingHandshakeResults = make(chan *handshaker.OutgoingHandshaker)
	t.incomingHandshakeResults = make(chan *handshaker.IncomingHandshaker)
	t.closeC = make(chan struct{})
	t.doneC = make(chan struct{})
	t.pid = sess.pid

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

func (t *Torrent) start() {
	go t.run()

	announceReq := tracker.AnnounceRequest{
		Infohash:   t.meta.Infohash,
		PeerID:     t.pid,
		Downloaded: 0,
		Uploaded:   0,
		Left:       0,
		Event:      tracker.TRACKER_STARTED,
		Port:       t.session.Port,
	}

	for _, tr := range t.trackers {
		res, err := tr.Announce(context.TODO(), announceReq)
		if err != nil {
			t.logger.Warn("torrent -> error in announcing to tracker", "tracker", tr.URL(), "error", err.Error())
			continue
		}
		for _, p := range res.Peers {
			go func() {
				conn, err := net.DialTimeout("tcp", p.String(), time.Second*5)
				if err != nil {
					t.logger.Warn("torrent -> error in connecting to peer", "address", p.Addr().String(), "error", err.Error())
					return
				}
				t.newConns <- conn
			}()
		}
	}
}

func (t *Torrent) run() {
	defer close(t.doneC)

	for {
		select {
		case <-t.closeC:
			for _, p := range t.peers {
				p.Stop()
			}
			for _, hs := range t.outgoingHandshakes {
				hs.Close()
			}

			announceStop := tracker.AnnounceRequest{
				Infohash:   t.meta.Infohash,
				PeerID:     t.pid,
				Downloaded: 0,
				Uploaded:   0,
				Left:       0,
				Event:      tracker.TRACKER_STOPPED,
				Port:       t.session.Port,
			}
			for _, tr := range t.trackers {
				tr.Announce(context.TODO(), announceStop)
			}
			t.logger.Info("torrent stopped")
			return
		case conn := <-t.newConns:
			{
				hs := handshaker.NewOutgoingHandshaker(conn)
				t.outgoingHandshakes = append(t.outgoingHandshakes, hs)
				go hs.Run(t.outgoingHandshakeResults, t.pid, t.meta.Infohash, [8]byte{0}, time.Second*2)
				t.logger.Info("torrent -> started handshaker for connection", "address", conn.RemoteAddr().String())
			}
		case res := <-t.outgoingHandshakeResults:
			{
				t.outgoingHandshakes = util.Remove(t.outgoingHandshakes, res, func(e1, e2 *handshaker.OutgoingHandshaker) bool { return e1 == e2 })
				if res.Error != nil {
					t.logger.Warn("torrent -> error during handshake", "address", res.Conn.RemoteAddr().String(), "error", res.Error.Error())
					continue
				}
				t.logger.Info("torrent -> connected to peer", "peer", string(res.PeerID[:]), "peers", len(t.peers))
				pe := peer.New(t.logger, res.PeerID, res.Conn)
				t.peers = append(t.peers, pe)
				go pe.Run(t.peerMessages, t.disconnectedPeers)
			}
		case res := <-t.incomingHandshakeResults:
			{
				if res.Error != nil {
					t.logger.Warn("torrent -> error during incoming handshake", "address", res.Conn.RemoteAddr().String(), "error", res.Error.Error())
					continue
				}
				t.logger.Info("torrent -> peer connected to us", "peer", string(res.PeerID[:]), "peers", len(t.peers))
				pe := peer.New(t.logger, res.PeerID, res.Conn)
				t.peers = append(t.peers, pe)
				go pe.Run(t.peerMessages, t.disconnectedPeers)
			}
		case pe := <-t.disconnectedPeers:
			{
				t.peers = util.Remove(t.peers, pe, func(e1, e2 *peer.Peer) bool { return e1 == e2 })
				t.logger.Info("torrent -> peer disconnected", "peer", string(pe.ID[:]), "peers", len(t.peers))
				pe.Stop()
			}
		case pe := <-t.peerMessages:
			{
				t.logger.Info("torrent -> received message from peer", "peer", string(pe.ID[:]), "message", pe.Message.ID().String())
			}
		}
	}
}

func (t *Torrent) stop() {
	close(t.closeC)
	<-t.doneC
}
