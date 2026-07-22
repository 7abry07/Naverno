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
	t.logger.Info("torrent -> peer disconnected", "Address", p.Addr().String(), "Peer", string(p.ID[:]), "Peer Count", len(t.peers))
	t.picker.OnPeerDisconnected(p)
	downloader, ok := t.downloaders[p]
	if ok {
		t.stalledDownloaders[downloader.Piece] = downloader
		t.picker.OnPieceStalled(downloader.Piece)
		downloader.OnPeerDisconnected()
		t.logger.Info("torrent -> downloader stalled", "Piece", downloader.Piece)
		delete(t.downloaders, p)
	}
	util.Remove(t.peers, p, func(e1, e2 *peer.Peer) bool { return e1 == e2 })
	p.Stop()
}

func (t *Torrent) handleNewConn(conn net.Conn) {
	hs := handshaker.NewOutgoingHandshaker(conn)
	t.outgoing = append(t.outgoing, hs)
	go hs.Run(t.outgoingResults, t.pid, t.meta.Infohash, t.extensions, time.Second*2)
	t.logger.Debug("torrent -> started handshaker for connection", "Address", conn.RemoteAddr().String())
}

func (t *Torrent) dial(peers []netip.AddrPort) {
	for _, a := range peers {
		go func() {
			conn, err := net.DialTimeout("tcp", a.String(), time.Second*5)
			if err != nil {
				t.logger.Debug("torrent -> error in connecting to remote peer", "Address", a.Addr().String(), "Error", err.Error())
				return
			}
			t.newConns <- conn
		}()
	}
}

func (t *Torrent) closePeer(p *peer.Peer) {
	delete(t.downloaders, p)
	p.Stop()
	util.Remove(t.peers, p, func(e1, e2 *peer.Peer) bool { return e1 == e2 })

}

func (t *Torrent) closePeers() {
	for _, p := range t.peers {
		p.Stop()
	}
}
