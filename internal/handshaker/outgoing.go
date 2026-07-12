package handshaker

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"time"
)

type OutgoingHandshaker struct {
	conn       net.Conn
	pid        [20]byte
	ih         [20]byte
	extensions [8]byte
	timeout    time.Duration
}

func NewOutgoingHandshaker(conn net.Conn, pid [20]byte, ih [20]byte, extensions [8]byte, timeout time.Duration) *OutgoingHandshaker {
	return &OutgoingHandshaker{
		conn:       conn,
		pid:        pid,
		ih:         ih,
		extensions: extensions,
		timeout:    timeout,
	}
}

func (o *OutgoingHandshaker) Run(result chan<- HandshakedConn) {
	protocolStr := "BitTorrent protocol"
	handshakeLen := 49 + len(protocolStr)

	buf := make([]byte, handshakeLen)
	buf = append(buf, byte(len(protocolStr)))
	buf = append(buf, o.extensions[:]...)
	buf = append(buf, o.ih[:]...)
	buf = append(buf, o.pid[:]...)

	o.conn.SetDeadline(time.Now().Add(o.timeout))
	_, err := o.conn.Write(buf)
	if err != nil {
		result <- HandshakedConn{
			nil,
			[20]byte{},
			[8]byte{},
			err,
		}
	}

	readBuf := bytes.NewBuffer(make([]byte, handshakeLen))
	_, err = io.ReadFull(o.conn, readBuf.Bytes())
	if err != nil {
		result <- HandshakedConn{
			nil,
			[20]byte{},
			[8]byte{},
			err,
		}
	}

	pstrlen, _ := readBuf.ReadByte()
	pstr := make([]byte, len(protocolStr))
	extensions := [8]byte{}
	ih := [20]byte{}
	pid := [20]byte{}

	readBuf.Read(pstr)
	readBuf.Read(extensions[:])
	readBuf.Read(ih[:])
	readBuf.Read(pid[:])

	if pstrlen != 19 {
		result <- HandshakedConn{
			nil,
			[20]byte{},
			[8]byte{},
			fmt.Errorf("protocol string length is invalid"),
		}
	}

	if !bytes.Equal(pstr, []byte(protocolStr)) {
		result <- HandshakedConn{
			nil,
			[20]byte{},
			[8]byte{},
			fmt.Errorf("protocol string is invalid"),
		}
	}

	if !bytes.Equal(ih[:], o.ih[:]) {
		result <- HandshakedConn{
			nil,
			[20]byte{},
			[8]byte{},
			fmt.Errorf("info hash is not equal"),
		}
	}

	for i, b := range extensions {
		extensions[i] = o.extensions[i] & b
	}

	result <- HandshakedConn{
		o.conn,
		pid,
		extensions,
		nil,
	}
}
