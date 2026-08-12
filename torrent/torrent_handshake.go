package torrent

import (
	"Naverno/internal/handshaker"
	"Naverno/internal/peer"
	"net"
	"time"
)

func (t *Torrent) handleNewConn(conn net.Conn) {
	hs := handshaker.NewOutgoingHandshaker(conn)
	t.outgoing[hs] = struct{}{}
	go hs.Run(t.outgoingResults, t.session.pid, t.meta.Infohash, t.session.extensions, time.Second*2)
	t.logger.Debug("torrent -> started handshaker for connection", "Address", conn.RemoteAddr().String())
}

func (t *Torrent) handleIncomingResult(res *handshaker.IncomingHandshaker) {
	if res.Error != nil {
		t.logger.Debug("torrent -> error during handshake", "Address", res.Conn.RemoteAddr().String(), "Error", res.Error.Error())
		return
	}
	if _, ok := t.peers[res.PeerID]; ok {
		return
	}
	t.logger.Debug("torrent -> peer connected to us", "Peer", string(res.PeerID[:]), "Peer Count", len(t.peers))
	pe := peer.New(t.logger, time.Now(), res.Conn, res.PeerID, res.Extensions)
	t.peers[pe.ID] = pe
	go pe.Run(t.peerMessages, t.disconnectedPeers)
	pe.Bitfield(t.bitset.Bytes())
}

func (t *Torrent) handleOutgoingResult(res *handshaker.OutgoingHandshaker) {
	delete(t.outgoing, res)
	if res.Error != nil {
		t.logger.Debug("torrent -> error during handshake", "Address", res.Conn.RemoteAddr().String(), "Error", res.Error.Error())
		return
	}
	if _, ok := t.peers[res.PeerID]; ok {
		return
	}
	t.logger.Debug("torrent -> connected to peer", "Peer", string(res.PeerID[:]), "Peer Count", len(t.peers))
	pe := peer.New(t.logger, time.Now(), res.Conn, res.PeerID, res.Extensions)
	t.peers[pe.ID] = pe
	go pe.Run(t.peerMessages, t.disconnectedPeers)
	pe.Bitfield(t.bitset.Bytes())
}

func (t *Torrent) closeHandshakes() {
	for hs := range t.outgoing {
		hs.Close()
	}
}
