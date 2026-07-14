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

	closeC chan struct{}
	doneC  chan struct{}
}

func New(conn net.Conn) *Reader {
	return &Reader{
		conn:     conn,
		messages: make(chan peerprotocol.Message),
		fatal:    make(chan error),
		closeC:   make(chan struct{}),
		doneC:    make(chan struct{}),
	}
}

func (r *Reader) Run() {
	go r.listen()
}

func (r *Reader) Messages() <-chan peerprotocol.Message {
	return r.messages
}

func (r *Reader) Close() {
	close(r.closeC)
	<-r.doneC
}

func (r *Reader) Error() <-chan error {
	return r.fatal
}

func (r *Reader) listen() {
	defer close(r.doneC)
	for {
		select {
		case <-r.closeC:
			return
		default:
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
}
