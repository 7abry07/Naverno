package writer

import (
	"Naverno/internal/peerprotocol"
	"Naverno/internal/util"
	"log/slog"
	"net"
)

type Writer struct {
	logger   *slog.Logger
	conn     net.Conn
	messages chan peerprotocol.Message
	fatal    chan error

	closeC chan struct{}
	doneC  chan struct{}
}

func New(logger *slog.Logger, conn net.Conn) *Writer {
	if logger == nil {
		panic("passed nil logger to peer writer")
	}

	return &Writer{
		logger:   logger,
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
			w.logger.Debug("writer -> wrote message", "Remote", w.conn.RemoteAddr().String(), "Message", mess.ID().String())
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
	select {
	case w.messages <- mess:
	case <-w.closeC:
	}
}
