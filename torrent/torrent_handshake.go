package torrent

import (
	"Naverno/internal/handshaker"
	"Naverno/internal/peer"
	"Naverno/internal/util"
)

func (t *Torrent) handleIncoming(res *handshaker.IncomingHandshaker) {
	if res.Error != nil {
		t.logger.Warn("torrent -> error during handshake", "Address", res.Conn.RemoteAddr().String(), "Error", res.Error.Error())
		return
	}
	t.logger.Info("torrent -> peer connected to us", "Peer", string(res.PeerID[:]), "Peer Count", len(t.peers))
	pe := peer.New(t.logger, res.Conn, res.PeerID, res.Extensions)
	t.peers = append(t.peers, pe)
	go pe.Run(t.peerMessages, t.disconnectedPeers)
}

func (t *Torrent) handleOutgoing(res *handshaker.OutgoingHandshaker) {
	t.outgoingHandshakes = util.Remove(t.outgoingHandshakes, res, func(e1, e2 *handshaker.OutgoingHandshaker) bool { return e1 == e2 })
	if res.Error != nil {
		t.logger.Warn("torrent -> error during handshake", "Address", res.Conn.RemoteAddr().String(), "Error", res.Error.Error())
		return
	}
	t.logger.Info("torrent -> connected to peer", "Peer", string(res.PeerID[:]), "Peer Count", len(t.peers))
	pe := peer.New(t.logger, res.Conn, res.PeerID, res.Extensions)
	t.peers = append(t.peers, pe)
	go pe.Run(t.peerMessages, t.disconnectedPeers)
}

func (t *Torrent) CloseHandshakes() {
	for _, hs := range t.outgoingHandshakes {
		hs.Close()
	}
}
