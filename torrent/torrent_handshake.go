package torrent

import (
	"Naverno/internal/handshaker"
	"Naverno/internal/peer"
	"Naverno/internal/util"
)

func (t *Torrent) handleIncomingResult(res *handshaker.IncomingHandshaker) {
	if res.Error != nil {
		t.logger.Debug("torrent -> error during handshake", "Address", res.Conn.RemoteAddr().String(), "Error", res.Error.Error())
		return
	}
	t.logger.Debug("torrent -> peer connected to us", "Peer", string(res.PeerID[:]), "Peer Count", len(t.peers))
	pe := peer.New(t.logger, res.Conn, res.PeerID, res.Extensions)
	t.peers = append(t.peers, pe)
	go pe.Run(t.peerMessages, t.disconnectedPeers)
	pe.Bitfield(util.BitsetToBytes(t.pieces))
}

func (t *Torrent) handleOutgoingResult(res *handshaker.OutgoingHandshaker) {
	t.outgoing = util.Remove(t.outgoing, res, func(e1, e2 *handshaker.OutgoingHandshaker) bool { return e1 == e2 })
	if res.Error != nil {
		t.logger.Debug("torrent -> error during handshake", "Address", res.Conn.RemoteAddr().String(), "Error", res.Error.Error())
		return
	}
	t.logger.Debug("torrent -> connected to peer", "Peer", string(res.PeerID[:]), "Peer Count", len(t.peers))
	pe := peer.New(t.logger, res.Conn, res.PeerID, res.Extensions)
	t.peers = append(t.peers, pe)
	go pe.Run(t.peerMessages, t.disconnectedPeers)
	pe.Bitfield(util.BitsetToBytes(t.pieces))
}

func (t *Torrent) closeHandshakes() {
	for _, hs := range t.outgoing {
		hs.Close()
	}
}
