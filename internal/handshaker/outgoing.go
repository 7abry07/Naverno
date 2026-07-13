package handshaker

import (
	"bytes"
	"fmt"
	"net"
	"time"
)

type OutgoingHandshaker struct {
	conn       net.Conn
	pid        [20]byte
	infohash   [20]byte
	extensions [8]byte
	timeout    time.Duration
}

func NewOutgoingHandshaker(conn net.Conn, pid [20]byte, ih [20]byte, extensions [8]byte, timeout time.Duration) *OutgoingHandshaker {
	return &OutgoingHandshaker{
		conn:       conn,
		pid:        pid,
		infohash:   ih,
		extensions: extensions,
		timeout:    timeout,
	}
}

func (o *OutgoingHandshaker) Run(result chan<- HandshakeResult) {
	hsResult := HandshakeResult{}
	remoteHs := Handshake{}
	hs := Handshake{
		Extensions: o.extensions,
		InfoHash:   o.infohash,
		PeerID:     o.pid,
	}

	o.conn.SetDeadline(time.Now().Add(o.timeout))
	_, err := o.conn.Write(hs.Marshal())
	if err != nil {
		hsResult.Error = err
		result <- hsResult
		return
	}

	err = remoteHs.Unmarshal(o.conn)
	if err != nil {
		hsResult.Error = err
		result <- hsResult
		return
	}

	if !bytes.Equal(remoteHs.InfoHash[:], o.infohash[:]) {
		hsResult.Error = fmt.Errorf("info hash is not equal")
		result <- hsResult
		return
	}

	for i, b := range remoteHs.Extensions {
		remoteHs.Extensions[i] = o.extensions[i] & b
	}

	hsResult.Conn = o.conn
	hsResult.PeerID = remoteHs.PeerID
	hsResult.Extensions = remoteHs.Extensions
	result <- hsResult
}
