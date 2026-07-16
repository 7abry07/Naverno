package handshaker

import (
	"Naverno/internal/util"
	"fmt"
	"net"
	"time"
)

type IncomingHandshaker struct {
	Conn       net.Conn
	PeerID     [20]byte
	InfoHash   [20]byte
	Extensions [8]byte
	Error      error

	closeC chan struct{}
	doneC  chan struct{}
}

func NewIncomingHandshaker(conn net.Conn) *IncomingHandshaker {
	return &IncomingHandshaker{
		Conn:   conn,
		closeC: make(chan struct{}),
		doneC:  make(chan struct{}),
	}
}

func (i *IncomingHandshaker) Run(result chan<- *IncomingHandshaker, validInfoHash func([20]byte) bool, pid [20]byte, extensions [8]byte, timeout time.Duration) {
	defer i.Conn.SetDeadline(time.Time{})
	defer close(i.doneC)
	defer func() {
		select {
		case result <- i:
		case <-i.closeC:
			i.Conn.Close()
		}
	}()
	i.Conn.SetDeadline(time.Now().Add(timeout))

	remoteHs := Handshake{}
	err := remoteHs.Unmarshal(i.Conn)
	if err != nil {
		i.Error = err
		return
	}

	if !validInfoHash(remoteHs.InfoHash) {
		i.Error = fmt.Errorf("infohash isn't valid")
		return
	}

	hs := Handshake{
		PeerID:     pid,
		InfoHash:   remoteHs.InfoHash,
		Extensions: extensions,
	}

	for i, b := range remoteHs.Extensions {
		remoteHs.Extensions[i] = extensions[i] & b
	}

	err = util.WriteFull(i.Conn, hs.Marshal())
	if err != nil {
		i.Error = err
		return
	}

	i.PeerID = remoteHs.PeerID
	i.InfoHash = remoteHs.InfoHash
	i.Extensions = remoteHs.Extensions
}

func (i *IncomingHandshaker) Close() {
	close(i.closeC)
	<-i.doneC
}
