package torrent

import (
	"Naverno/internal/handshaker"
	"Naverno/internal/peer"
	"Naverno/internal/util"
	"net"
	"net/netip"
	"time"
)

func (t *Torrent) handleDisconnected(p *peer.Peer) {
	t.peers = util.Remove(t.peers, p, func(e1, e2 *peer.Peer) bool { return e1 == e2 })
	t.logger.Info("torrent -> peer disconnected", "Peer", string(p.ID[:]), "Peer Count", len(t.peers))
	p.Stop()
}

func (t *Torrent) handleNewConn(conn net.Conn) {
	hs := handshaker.NewOutgoingHandshaker(conn)
	t.outgoing = append(t.outgoing, hs)
	go hs.Run(t.outgoingResults, t.pid, t.meta.Infohash, t.extensions, time.Second*2)
	t.logger.Info("torrent -> started handshaker for connection", "Address", conn.RemoteAddr().String())
}

func (t *Torrent) Dial(peers []netip.AddrPort) {
	for _, a := range peers {
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

func (t *Torrent) closePeers() {
	for _, p := range t.peers {
		p.Stop()
	}
}
