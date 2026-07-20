package writer

import (
	"Naverno/internal/peerprotocol"
	"Naverno/internal/util"
	"net"
)

type Writer struct {
	conn     net.Conn
	messages chan peerprotocol.Message
	fatal    chan error

	closeC chan struct{}
	doneC  chan struct{}
}

func New(conn net.Conn) *Writer {
	return &Writer{
		conn:     conn,
		messages: make(chan peerprotocol.Message),
		fatal:    make(chan error),
		closeC:   make(chan struct{}),
		doneC:    make(chan struct{}),
	}
}

func (w *Writer) Run() {
	defer close(w.doneC)
	for {
		select {
		case <-w.closeC:
			return
		case mess := <-w.messages:
			err := util.WriteFull(w.conn, mess.Marshal())
			if err != nil {
				select {
				case <-w.closeC:
				case w.fatal <- err:
				}
				return
			}
		}
	}
}

func (w *Writer) Close() {
	close(w.closeC)
	<-w.doneC
}

func (w *Writer) Error() <-chan error {
	return w.fatal
}

func (w *Writer) Write(mess peerprotocol.Message) {
	w.messages <- mess
}
