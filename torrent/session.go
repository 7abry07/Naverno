package torrent

import (
	"Naverno/internal/handshaker"
	"Naverno/internal/metadata"
	"Naverno/internal/peer"
	"Naverno/internal/trackermanager"
	"Naverno/internal/util"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Session struct {
	torrents    map[[20]byte]*Torrent
	torrentsMut sync.Mutex

	trackerManager *trackermanager.TrackerManager
	logger         *slog.Logger
	listener       net.Listener

	pid [20]byte

	incomingConns             chan net.Conn
	incomingHandshakes        []*handshaker.IncomingHandshaker
	incomingHandshakesResults chan *handshaker.IncomingHandshaker

	Port uint16

	errorC chan error
	doneC  chan struct{}
}

func NewSession(logger *slog.Logger) *Session {
	if logger == nil {
		panic("cannot pass nil logger to session")
	}

	s := Session{}

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	addrport := listener.Addr().String()
	portStartIdx := strings.LastIndexByte(addrport, ':')
	portstr := addrport[portStartIdx+1:]
	port, err := strconv.ParseUint(portstr, 10, 16)
	if err != nil {
		panic(err)
	}

	s.listener = listener
	s.doneC = make(chan struct{})
	s.torrents = make(map[[20]byte]*Torrent)
	s.torrentsMut = sync.Mutex{}
	s.Port = uint16(port)
	s.trackerManager = trackermanager.New(logger)
	s.incomingHandshakes = []*handshaker.IncomingHandshaker{}
	s.incomingHandshakesResults = make(chan *handshaker.IncomingHandshaker)
	s.logger = logger
	s.errorC = make(chan error)
	s.pid = peer.GenerateRandomID()

	return &s
}

func (s *Session) NewTorrentFromFile(path string) (*Torrent, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening torrent file -> %v", err)
	}

	meta, err := metadata.New(file)
	if err != nil {
		return nil, fmt.Errorf("error creating torrent metadata -> %v", err)
	}

	t, err := newTorrentFromMetadata(s, meta)
	if err != nil {
		return nil, err
	}

	s.torrents[t.meta.Infohash] = t

	return t, nil
}

func (s *Session) Run() error {
	defer close(s.doneC)
	go s.listen()
	for _, torr := range s.torrents {
		torr.start()
	}

	for {
		select {
		case err := <-s.errorC:
			{
				s.listener.Close()
				for _, torr := range s.torrents {
					torr.stop()
				}
				s.logger.Info("session stopped")
				return err
			}
		case conn := <-s.incomingConns:
			{
				hs := handshaker.NewIncomingHandshaker(conn)
				s.incomingHandshakes = append(s.incomingHandshakes, hs)
				go hs.Run(s.incomingHandshakesResults, s.checkInfoHash, s.pid, [8]byte{}, time.Second*5)

			}
		case res := <-s.incomingHandshakesResults:
			{
				s.torrentsMut.Lock()
				s.incomingHandshakes = util.Remove(s.incomingHandshakes, res, func(e1, e2 *handshaker.IncomingHandshaker) bool { return e1 == e2 })
				torr, ok := s.torrents[res.InfoHash]
				if !ok {
					panic("didn't find torrent in map from the handshaker infohash")
				}
				torr.incomingHandshakeResults <- res
				s.torrentsMut.Unlock()
			}
		}
	}
}

func (s *Session) Stop() {
	s.errorC <- nil
	<-s.doneC
}

func (s *Session) checkInfoHash(ih [20]byte) bool {
	defer s.torrentsMut.Unlock()
	s.torrentsMut.Lock()
	_, ok := s.torrents[ih]
	return ok
}

func (s *Session) listen() {
	conn, err := s.listener.Accept()
	if err != nil {
		s.errorC <- err
		return
	}
	s.incomingConns <- conn
}
