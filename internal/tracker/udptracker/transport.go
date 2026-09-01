package udptracker

import (
	"Naverno/internal/util"
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"net"
	"net/url"
	"sync"
)

var (
	TransportClosedErr error = errors.New("closed udp transport")
)

type Request interface {
	Encode(uint32) []byte
}

type Connection struct {
	*net.UDPConn
	requests []uint32
}

type UDPTransportRequest struct {
	ctx      context.Context
	url      url.URL
	request  Request
	response chan any
}

func NewUDPTransportRequest(ctx context.Context, url url.URL, request Request) *UDPTransportRequest {
	return &UDPTransportRequest{
		ctx:      ctx,
		url:      url,
		request:  request,
		response: make(chan any),
	}
}

type UDPTransport struct {
	connections    map[string]*Connection
	pending        map[uint32]*UDPTransportRequest
	connectionsMut sync.Mutex
	pendingMut     sync.Mutex
	req            chan *UDPTransportRequest

	closeC chan struct{}
	doneC  chan struct{}
}

func NewUDPTransport() *UDPTransport {
	return &UDPTransport{
		connections: make(map[string]*Connection),
		pending:     make(map[uint32]*UDPTransportRequest),
		req:         make(chan *UDPTransportRequest),
		closeC:      make(chan struct{}),
		doneC:       make(chan struct{}),
	}
}

func (t *UDPTransport) Do(req *UDPTransportRequest) (any, error) {
	select {
	case <-req.ctx.Done():
		return nil, req.ctx.Err()
	case t.req <- req:
	case <-t.doneC:
		return nil, TransportClosedErr
	}
	var res any
	select {
	case res = <-req.response:
	case <-req.ctx.Done():
		return nil, req.ctx.Err()
	}
	switch res := res.(type) {
	case connectResponse:
		switch req.request.(type) {
		case connectRequest:
			return res, nil
		default:
			return nil, fmt.Errorf("response-request mismatch")
		}
	case announceResponse:
		switch req.request.(type) {
		case announceRequest:
			return res, nil
		default:
			return nil, fmt.Errorf("response-request mismatch")
		}
	case errorResponse:
		return nil, fmt.Errorf("tracker error -> %v", res.message)
	case error:
		return nil, res
	default:
		panic("the response type is unknown")
	}
}

func (t *UDPTransport) getConnection(ctx context.Context, host string) (*Connection, error) {
	t.connectionsMut.Lock()
	defer t.connectionsMut.Unlock()
	conn, ok := t.connections[host]
	if !ok {
		dialer := net.Dialer{}
		c, err := dialer.DialContext(ctx, "udp", host)
		if err != nil {
			return nil, fmt.Errorf("dial error -> %v", err)
		}
		conn = &Connection{UDPConn: c.(*net.UDPConn), requests: []uint32{}}
		t.connections[host] = conn
	}
	return conn, nil
}

func (t *UDPTransport) Run() {
	defer close(t.doneC)

	for {
		select {
		case <-t.closeC:
			return
		case r := <-t.req:
			done := make(chan struct{})

			conn, err := t.getConnection(r.ctx, r.url.Host)
			if err != nil {
				r.response <- err
				continue
			}
			go t.readLoopConn(conn)

			go func() {
				select {
				case <-done:
				case <-r.ctx.Done():
					conn.Close()
				}
			}()

			transactionID := rand.Uint32()
			req := r.request.Encode(transactionID)
			err = util.WriteFull(conn, req)
			conn.requests = append(conn.requests, transactionID)
			if err != nil {
				t.deleteConnection(conn)
				continue
			}
			t.pendingMut.Lock()
			t.pending[transactionID] = r
			t.pendingMut.Unlock()
			close(done)
		}
	}
}

func (t *UDPTransport) Close() {
	close(t.closeC)
	for _, conn := range t.connections {
		conn.Close()
	}
	<-t.doneC
}

func (t *UDPTransport) deleteConnection(conn *Connection) {
	t.pendingMut.Lock()
	t.connectionsMut.Lock()
	defer t.connectionsMut.Unlock()
	defer t.pendingMut.Unlock()
	delete(t.connections, conn.RemoteAddr().String())
	conn.Close()
	for _, r := range conn.requests {
		pending, ok := t.pending[r]
		if !ok {
			panic("found pending request that wasn't pending")
		}
		delete(t.pending, r)
		close(pending.response)
	}
}

func (t *UDPTransport) readLoopConn(conn *Connection) {
	for {
		buf := make([]byte, 65535)
		read, err := conn.Read(buf)
		if err != nil {
			t.deleteConnection(conn)
			return
		}

		buf = buf[:read]

		temp := bytes.NewBuffer(buf)
		a, tid, err := getResponseInfo(temp)
		if err != nil {
			continue
		}
		t.pendingMut.Lock()
		pending, ok := t.pending[tid]
		if !ok {
			t.pendingMut.Unlock()
			continue
		}
		delete(t.pending, tid)
		t.pendingMut.Unlock()

		connReqs := []uint32{}
		for _, r := range conn.requests {
			if r == tid {
				continue
			}
			connReqs = append(connReqs, r)
		}
		conn.requests = connReqs

		var res any
		switch a {
		case action_connect:
			connect := connectResponse{}
			err := connect.decode(temp)
			if err != nil {
				res = err
				continue
			}
			res = connect
		case action_announce:
			ann := announceResponse{}
			err := ann.decode(temp)
			if err != nil {
				res = err
				continue
			}
			res = ann
			pending.response <- ann
		case action_error:
			errResp := errorResponse{}
			err := errResp.decode(temp)
			if err != nil {
				res = err
				continue
			}
			res = errResp
		default:
			res = fmt.Errorf("unrecognized action")
			continue
		}
		select {
		case pending.response <- res:
		case <-pending.ctx.Done():
		case <-t.closeC:
		}
	}
}
