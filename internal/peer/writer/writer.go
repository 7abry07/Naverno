package writer

import (
	"Naverno/internal/blockingqueue"
	"Naverno/internal/peerprotocol"
	"Naverno/internal/util"
	"log/slog"
	"net"
	"time"
)

type Writer struct {
	logger *slog.Logger
	conn   net.Conn
	queue  *blockingqueue.Queue[peerprotocol.Message]
	fatal  chan error

	closeC chan struct{}
	doneC  chan struct{}
}

func New(logger *slog.Logger, conn net.Conn) *Writer {
	if logger == nil {
		panic("passed nil logger to peer writer")
	}

	writer := &Writer{
		logger: logger,
		conn:   conn,
		queue:  blockingqueue.New([]peerprotocol.Message{}),
		fatal:  make(chan error),
		closeC: make(chan struct{}),
		doneC:  make(chan struct{}),
	}
	return writer
}

func (w *Writer) Run() {
	defer close(w.doneC)
	for {
		mess, ok := w.queue.Pop()
		if !ok {
			return
		}
		select {
		case <-w.closeC:
			return
		default:
		}

		w.conn.SetWriteDeadline(time.Now().Add(time.Second * 30))
		err := util.WriteFull(w.conn, mess.Marshal())
		w.conn.SetWriteDeadline(time.Time{})
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

func (w *Writer) Close() {
	close(w.closeC)
	w.queue.Stop()
	<-w.doneC
}

func (w *Writer) Error() <-chan error {
	return w.fatal
}

func (w *Writer) Write(mess peerprotocol.Message) {
	select {
	case <-w.closeC:
	default:
	}
	w.queue.Push(mess)
}
