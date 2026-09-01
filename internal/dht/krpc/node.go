package krpc

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	NodeClosed error = errors.New("closed krpc node")
)

type Handler func(w *ResponseWriter, q Query)
type ResponseWriter struct {
	values  map[string]any
	errCode ErrorCode
	errMess string
}

func (w *ResponseWriter) WriteValue(key string, val any) {
	if w.values == nil {
		w.values = make(map[string]any)
	}
	w.values[key] = val
}

func (w *ResponseWriter) WriteError(code ErrorCode, mess string) {
	w.errCode = code
	w.errMess = mess
}

type incomingQuery struct {
	*net.UDPAddr
	tid uint16
	Query
}

type outgoingQuery struct {
	ctx context.Context
	*net.UDPAddr
	Query
	res chan any
}

type Node struct {
	handlers sync.Map
	pending  sync.Map
	conn     *net.UDPConn

	queriesOut chan *outgoingQuery
	queriesIn  chan *incomingQuery

	port uint16

	err    chan error
	closeC chan struct{}
	doneC  chan struct{}
}

func New() (*Node, error) {
	addr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}

	addrport := conn.LocalAddr().String()
	portStartIdx := strings.LastIndexByte(addrport, ':')
	portstr := addrport[portStartIdx+1:]
	port, err := strconv.ParseUint(portstr, 10, 16)
	if err != nil {
		return nil, err
	}

	return &Node{
		conn:       conn,
		port:       uint16(port),
		queriesIn:  make(chan *incomingQuery),
		queriesOut: make(chan *outgoingQuery),
		closeC:     make(chan struct{}),
		doneC:      make(chan struct{}),
		err:        make(chan error),
	}, nil
}

func (n *Node) Query(ctx context.Context, addr string, query Query) (Message, error) {
	resolved, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	q := &outgoingQuery{
		ctx:     ctx,
		Query:   query,
		UDPAddr: resolved,
		res:     make(chan any),
	}

	select {
	case <-q.ctx.Done():
		return nil, q.ctx.Err()
	case n.queriesOut <- q:
	case <-n.doneC:
		return nil, NodeClosed
	}

	var res any
	select {
	case res = <-q.res:
	case <-q.ctx.Done():
		return nil, q.ctx.Err()
	}

	switch res := res.(type) {
	case Response:
		return res, nil
	case Error:
		return res, nil
	case error:
		return nil, res
	default:
		panic("impossible type returned")
	}
}

func (n *Node) Run() {
	defer close(n.doneC)
	go n.listen()

	for {
		select {
		case <-n.closeC:
			return
		case q := <-n.queriesIn:
			val, ok := n.handlers.Load(q.Name)
			if !ok {
				e := Error{
					Code:    ServerErrorCode,
					Message: "can't handle query",
				}
				n.conn.WriteToUDP(e.Marshal(q.tid), q.UDPAddr)
				continue
			}
			handler := val.(Handler)
			w := &ResponseWriter{}
			handler(w, q.Query)

			if w.errCode < 0 {
				e := Error{
					Code:    w.errCode,
					Message: w.errMess,
				}
				n.conn.WriteToUDP(e.Marshal(q.tid), q.UDPAddr)
				continue
			}

			r := Response{Values: w.values}
			n.conn.WriteToUDP(r.Marshal(q.tid), q.UDPAddr)
		case q := <-n.queriesOut:
			done := make(chan struct{})

			go func() {
				select {
				case <-done:
				case <-q.ctx.Done():
					n.conn.SetWriteDeadline(time.Now())
				}
			}()

			transactionID := uint16(rand.Uint32())
			_, err := n.conn.WriteToUDP(q.Marshal(transactionID), q.UDPAddr)
			n.conn.SetWriteDeadline(time.Time{})
			close(done)
			if err == nil {
				n.pending.Store(transactionID, q)
			}
		}
	}
}

func (n *Node) RegisterHandle(query string, handler Handler) { n.handlers.Store(query, handler) }
func (n *Node) Port() uint16                                 { return n.port }
func (n *Node) Error() <-chan error                          { return n.err }

func (n *Node) Close() {
	close(n.closeC)
	n.pending.Range(func(key, value any) bool {
		pending := value.(*outgoingQuery)
		select {
		case pending.res <- NodeClosed:
		case <-pending.ctx.Done():
		}
		return true
	})
	<-n.doneC
	close(n.err)
}

func (n *Node) listen() {
	for {
		buf := make([]byte, 65535)
		read, addr, err := n.conn.ReadFromUDP(buf)
		if err != nil {
			n.err <- err
			close(n.err)
			return
		}

		buf = buf[:read]

		var res any
		mess, tid, err := Decode(buf)
		if tid == 0 {
			continue
		}

		val, ok := n.pending.Load(tid)
		if !ok {
			if err != nil {
				switch err := err.(type) {
				case Error:
					n.conn.WriteToUDP(err.Marshal(tid), addr)
				}
			} else {
				switch mess := mess.(type) {
				case Query:
					q := &incomingQuery{
						UDPAddr: addr,
						tid:     tid,
						Query:   mess,
					}
					select {
					case n.queriesIn <- q:
					case <-n.closeC:
					}
				}
			}
			continue
		}

		pending := val.(*outgoingQuery)
		if err != nil {
			res = err
		} else {
			switch mess.(type) {
			case Response:
				res = mess
			case Error:
				res = mess
			case Query:
				res = fmt.Errorf("query in response to a query")
			default:
				panic("impossible type returned")
			}
		}

		select {
		case pending.res <- res:
		case <-pending.ctx.Done():
		case <-n.closeC:
		}

		n.pending.Delete(tid)
	}
}
