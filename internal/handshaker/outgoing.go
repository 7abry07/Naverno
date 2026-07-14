package handshaker

import (
	"bytes"
	"fmt"
	"net"
	"time"
)

type OutgoingHandshaker struct {
	Conn       net.Conn
	PeerID     [20]byte
	Extensions [8]byte
	Error      error

	closeC chan struct{}
	doneC  chan struct{}
}

func NewOutgoingHandshaker(conn net.Conn) *OutgoingHandshaker {
	return &OutgoingHandshaker{
		Conn:   conn,
		closeC: make(chan struct{}),
		doneC:  make(chan struct{}),
	}
}

func (o *OutgoingHandshaker) Run(result chan<- *OutgoingHandshaker, pid [20]byte, ih [20]byte, extensions [8]byte, timeout time.Duration) {
	defer close(o.doneC)
	defer func() {
		select {
		case result <- o:
		case <-o.closeC:
			o.Conn.Close()
		}
	}()

	remoteHs := Handshake{}
	hs := Handshake{
		Extensions: extensions,
		InfoHash:   ih,
		PeerID:     pid,
	}

	o.Conn.SetDeadline(time.Now().Add(timeout))
	_, err := o.Conn.Write(hs.Marshal())
	if err != nil {
		o.Error = err
		return
	}

	err = remoteHs.Unmarshal(o.Conn)
	if err != nil {
		o.Error = err
		return
	}

	if !bytes.Equal(remoteHs.InfoHash[:], ih[:]) {
		o.Error = fmt.Errorf("info hash is not equal")
		return
	}

	for i, b := range remoteHs.Extensions {
		remoteHs.Extensions[i] = extensions[i] & b
	}

	o.PeerID = remoteHs.PeerID
	o.Extensions = remoteHs.Extensions
}

func (o *OutgoingHandshaker) Close() {
	close(o.closeC)
	<-o.doneC
}
