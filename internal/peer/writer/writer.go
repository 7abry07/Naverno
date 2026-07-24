package writer

import (
	"Naverno/internal/peerprotocol"
	"Naverno/internal/util"
	"log/slog"
	"net"
	"sync"
	"time"
)

type Writer struct {
	logger *slog.Logger
	conn   net.Conn
	cond   *sync.Cond
	queue  []peerprotocol.Message
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
		cond:   sync.NewCond(&sync.Mutex{}),
		queue:  []peerprotocol.Message{},
		fatal:  make(chan error),
		closeC: make(chan struct{}),
		doneC:  make(chan struct{}),
	}
	return writer
}

func (w *Writer) Run() {
	w.cond.L.Lock()
	defer close(w.doneC)
	for {
		if len(w.queue) == 0 {
			w.cond.Wait()
		}
		select {
		case <-w.closeC:
			return
		default:
		}

		mess := w.queue[0]
		w.queue = w.queue[1:]
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
	w.cond.Signal()
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
	w.cond.L.Lock()
	w.queue = append(w.queue, mess)
	w.cond.Signal()
	w.cond.L.Unlock()
}
