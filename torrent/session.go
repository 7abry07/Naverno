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
	pid  [20]byte
	port uint16

	listener           net.Listener
	trackerManager     *trackermanager.TrackerManager
	logger             *slog.Logger
	torrents           map[[20]byte]*Torrent
	torrentsMut        sync.Mutex
	incomingHandshakes []*handshaker.IncomingHandshaker

	newTorrent                chan *Torrent
	removeTorrent             chan *Torrent
	incomingConns             chan net.Conn
	incomingHandshakesResults chan *handshaker.IncomingHandshaker

	listenErr chan error
	closeC    chan struct{}
	doneC     chan struct{}
}

func StartSession(logger *slog.Logger) *Session {
	if logger == nil {
		panic("cannot pass nil logger to session")
	}

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

	s := Session{
		logger:                    logger,
		listener:                  listener,
		torrents:                  make(map[[20]byte]*Torrent),
		torrentsMut:               sync.Mutex{},
		port:                      uint16(port),
		pid:                       peer.GenerateRandomID(),
		trackerManager:            trackermanager.New(logger),
		incomingHandshakes:        []*handshaker.IncomingHandshaker{},
		newTorrent:                make(chan *Torrent),
		removeTorrent:             make(chan *Torrent),
		incomingHandshakesResults: make(chan *handshaker.IncomingHandshaker),
		listenErr:                 make(chan error),
		closeC:                    make(chan struct{}),
		doneC:                     make(chan struct{}),
	}

	go s.Run()

	return &s
}

func (s *Session) Run() {
	defer close(s.doneC)
	go s.listen()

	for {
		select {
		case <-s.closeC:
			{
				s.listener.Close()
				for _, torr := range s.torrents {
					torr.Stop()
				}
				for _, hs := range s.incomingHandshakes {
					hs.Close()
				}
				s.logger.Info("session stopped")
				return
			}
		case err := <-s.listenErr:
			{
				s.logger.Error("session -> error while listening for connection", "error", err.Error())
				close(s.closeC)
			}
		case t := <-s.newTorrent:
			{
				t.start()
				s.torrentsMut.Lock()
				s.torrents[t.meta.Infohash] = t
				s.torrentsMut.Unlock()
			}
		case t := <-s.removeTorrent:
			{
				t.Stop()
				s.torrentsMut.Lock()
				delete(s.torrents, t.meta.Infohash)
				s.torrentsMut.Unlock()
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
	close(s.closeC)
	<-s.doneC
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

	s.newTorrent <- t

	return t, nil
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
		select {
		case <-s.closeC:
		case s.listenErr <- err:
		}
		return
	}
	s.incomingConns <- conn
}
