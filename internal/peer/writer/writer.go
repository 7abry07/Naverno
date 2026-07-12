package writer

import (
	"Naverno/internal/peerprotocol"
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
		err := w.writeMessage(mess)
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

func (w *Writer) writeMessage(mess peerprotocol.Message) error {
	data := mess.Marshal()
	for len(data) > 0 {
		n, err := w.conn.Write(data)
		if err != nil {
			return err
		}
		data = data[n:]
	}
	return nil
}
