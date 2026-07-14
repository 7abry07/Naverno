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
}

func New(conn net.Conn) *Writer {
	return &Writer{
		conn:     conn,
		messages: make(chan peerprotocol.Message),
		fatal:    make(chan error),
	}
}

func (w *Writer) Run() {
	for mess := range w.messages {
		err := util.WriteFull(w.conn, mess.Marshal())
		if err != nil {
			w.fatal <- err
			return
		}
	}
}

func (w *Writer) Error() <-chan error {
	return w.fatal
}

func (w *Writer) Write(mess peerprotocol.Message) {
	w.messages <- mess
}
