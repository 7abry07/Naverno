package reader

import (
	"Naverno/internal/peerprotocol"
	"encoding/binary"
	"io"
	"net"
)

type Reader struct {
	conn     net.Conn
	messages chan peerprotocol.Message
	fatal    chan error
}

func New(conn net.Conn) *Reader {
	return &Reader{
		conn:     conn,
		messages: make(chan peerprotocol.Message),
		fatal:    make(chan error),
	}
}

func (r *Reader) Run() {
	go r.listen()
}

func (r *Reader) Messages() <-chan peerprotocol.Message {
	return r.messages
}

func (r *Reader) Error() <-chan error {
	return r.fatal
}

func (r *Reader) listen() {
	for {
		lengthBytes := make([]byte, 4)
		_, err := io.ReadFull(r.conn, lengthBytes)
		if err != nil {
			r.fatal <- err
			return
		}
		length := binary.BigEndian.Uint32(lengthBytes)

		messBytes := make([]byte, length)
		_, err = io.ReadFull(r.conn, messBytes)

		fullMess := []byte{}
		fullMess = append(fullMess, lengthBytes...)
		fullMess = append(fullMess, messBytes...)

		mess, err := peerprotocol.Decode(fullMess)
		if err != nil {
			r.fatal <- err
			return
		}

		r.messages <- mess
	}
}
