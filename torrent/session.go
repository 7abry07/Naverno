package torrent

import (
	"Naverno/internal/handshaker"
	"Naverno/internal/peer"
	"Naverno/internal/trackermanager"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"sync"
)

type Session struct {
	currentTid uint32
	pid        [20]byte
	port       uint16
	extensions [8]byte

	listener       net.Listener
	trackerManager *trackermanager.TrackerManager
	logger         *slog.Logger
	torrents       map[[20]byte]*Torrent
	torrentsMut    sync.Mutex
	incoming       []*handshaker.IncomingHandshaker

	newTorrent      chan *Torrent
	removeTorrent   chan *Torrent
	incomingConns   chan net.Conn
	incomingResults chan *handshaker.IncomingHandshaker

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
		logger:          logger,
		listener:        listener,
		torrents:        make(map[[20]byte]*Torrent),
		torrentsMut:     sync.Mutex{},
		port:            uint16(port),
		currentTid:      0,
		pid:             peer.GenerateRandomID(),
		extensions:      [8]byte{},
		trackerManager:  trackermanager.New(logger),
		incoming:        []*handshaker.IncomingHandshaker{},
		newTorrent:      make(chan *Torrent),
		removeTorrent:   make(chan *Torrent),
		incomingResults: make(chan *handshaker.IncomingHandshaker),
		listenErr:       make(chan error),
		closeC:          make(chan struct{}),
		doneC:           make(chan struct{}),
	}

	go s.Run()

	return &s
}

func (s *Session) Run() {

	defer close(s.doneC)
	defer s.listener.Close()
	defer s.trackerManager.Close()
	defer s.stopTorrents()
	defer s.stopHandshakes()
	defer s.logger.Info("session stopped")

	go s.listen()

	for {
		select {
		case <-s.closeC:
			return
		case err := <-s.listenErr:
			s.logger.Error("session -> error while listening for connection", "error", err.Error())
			close(s.closeC)
		case t := <-s.newTorrent:
			s.handleNewTorrent(t)
		case t := <-s.removeTorrent:
			s.handleRemoveTorrent(t)
		case conn := <-s.incomingConns:
			s.handleIncomingConn(conn)
		case res := <-s.incomingResults:
			s.handleIncomingResult(res)
		}
	}
}

func (s *Session) Stop() {
	close(s.closeC)
	<-s.doneC
}
