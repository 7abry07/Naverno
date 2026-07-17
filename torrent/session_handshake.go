package torrent

import (
	"Naverno/internal/handshaker"
	"Naverno/internal/util"
	"net"
	"time"
)

func (s *Session) handleIncomingConn(conn net.Conn) {
	s.logger.Info("session -> started handshaker for connection", "Address", conn.RemoteAddr().String())
	hs := handshaker.NewIncomingHandshaker(conn)
	s.incoming = append(s.incoming, hs)
	go hs.Run(s.incomingResults, s.checkInfoHash, s.pid, s.extensions, time.Second*5)
}

func (s *Session) handleIncomingResult(res *handshaker.IncomingHandshaker) {
	s.torrentsMut.Lock()
	defer s.torrentsMut.Unlock()
	s.incoming = util.Remove(s.incoming, res, func(e1, e2 *handshaker.IncomingHandshaker) bool { return e1 == e2 })
	torr, ok := s.torrents[res.InfoHash]
	if !ok {
		panic("didn't find torrent in map from the handshaker infohash")
	}
	torr.incomingResults <- res
}

func (s *Session) checkInfoHash(ih [20]byte) bool {
	defer s.torrentsMut.Unlock()
	s.torrentsMut.Lock()
	_, ok := s.torrents[ih]
	return ok
}

func (s *Session) stopHandshakes() {
	for _, hs := range s.incoming {
		hs.Close()
	}
}
