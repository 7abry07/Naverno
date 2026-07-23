package reader

import (
	"Naverno/internal/peerprotocol"
	"encoding/binary"
	"io"
	"log/slog"
	"net"
)

type Reader struct {
	logger   *slog.Logger
	conn     net.Conn
	messages chan peerprotocol.Message
	fatal    chan error

	closeC chan struct{}
	doneC  chan struct{}
}

func New(logger *slog.Logger, conn net.Conn) *Reader {
	if logger == nil {
		panic("passed nil logger to peer writer")
	}

	return &Reader{
		logger:   logger,
		conn:     conn,
		messages: make(chan peerprotocol.Message),
		fatal:    make(chan error),
		closeC:   make(chan struct{}),
		doneC:    make(chan struct{}),
	}
}

func (r *Reader) Run() {
	defer close(r.doneC)
	defer close(r.messages)
	defer close(r.fatal)
	for {
		lengthBytes := make([]byte, 4)
		_, err := io.ReadFull(r.conn, lengthBytes)
		if err != nil {
			select {
			case <-r.closeC:
			case r.fatal <- err:
			}
			return
		}
		length := binary.BigEndian.Uint32(lengthBytes)

		messBytes := make([]byte, length)
		_, err = io.ReadFull(r.conn, messBytes)
		if err != nil {
			select {
			case <-r.closeC:
			case r.fatal <- err:
			}
			return
		}

		fullMess := []byte{}
		fullMess = append(fullMess, lengthBytes...)
		fullMess = append(fullMess, messBytes...)

		mess, err := peerprotocol.Decode(fullMess)
		if err != nil {
			select {
			case <-r.closeC:
			case r.fatal <- err:
			}
			return
		}
		select {
		case r.messages <- mess:
			r.logger.Debug("reader -> message read", "Remote", r.conn.RemoteAddr().String(), "Message", mess.ID().String())
		case <-r.closeC:
		}

	}
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
