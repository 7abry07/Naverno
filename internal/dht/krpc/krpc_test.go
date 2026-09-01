package krpc_test

import (
	"Naverno/internal/dht/krpc"
	"context"
	"fmt"
	"maps"
	"net"
	"testing"
	"time"
)

func TestNodeQueryRecv(t *testing.T) {
	node, err := krpc.New()
	if err != nil {
		t.Fatalf("unexpected error -> %v", err)
	}

	node.RegisterHandle("hello", func(w *krpc.ResponseWriter, q krpc.Query) {
		value, ok := q.Parameters["value"]
		if !ok {
			w.WriteProtocolError("missing \"value\" key")
		}
		w.WriteValue("value", value)
	})

	go node.Run()
	defer node.Close()

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%v", node.Port()))
	if err != nil {
		t.Fatalf("unexpected error -> %v", err)
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		t.Fatalf("unexpected error -> %v", err)
	}

	q := krpc.Query{
		Name:       "hello",
		Parameters: map[string]any{"value": 5},
	}
	tid := uint16(47)
	_, err = conn.Write(q.Marshal(tid))
	if err != nil {
		t.Fatalf("unexpected error -> %v", err)
	}
	buf := make([]byte, 65535)
	read, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("unexpected error -> %v", err)
	}
	buf = buf[:read]

	//g
	res, messtid, err := krpc.Decode(buf)
	if err != nil {
		t.Fatalf("unexpected error -> %v", err)
	}

	if messtid != tid {
		t.Errorf("transaction id doesn't match-> %v", err)
	}

	switch res := res.(type) {
	case krpc.Error:
		t.Errorf("expecting response, got error -> code: %v | message : %v", res.Code(), res.Error())
	case krpc.Response:
		valif, ok := res.Values["value"]
		if !ok {
			t.Fatalf("missing value in response")
		}
		val := valif.(int64)
		if val != 5 {
			t.Fatalf("value in response doesn't match, got -> %v", val)
		}
	}
}

func TestNodeQuerySend(t *testing.T) {
	node, err := krpc.New()
	if err != nil {
		t.Fatalf("unexpected error -> %v", err)
	}

	go node.Run()

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%v", node.Port()))
	if err != nil {
		t.Fatalf("unexpected error -> %v", err)
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		t.Fatalf("unexpected error -> %v", err)
	}

	go func() {
		buf := make([]byte, 65535)
		read, err := conn.Read(buf)
		if err != nil {
			return
		}
		buf = buf[:read]

		var res krpc.Message
		mess, tid, err := krpc.Decode(buf)
		if err != nil && tid == 0 {
			return
		}

		switch e := err.(type) {
		case krpc.Error:
			res = e
		case nil:
		default:
			res = krpc.GenericError{err.Error()}
		}

		q := mess.(krpc.Query)
		resp := krpc.Response{
			Values: map[string]any{},
		}
		maps.Copy(resp.Values, q.Parameters)
		res = resp

		_, err = conn.Write(res.Marshal(tid))
		if err != nil {
			return
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	q := krpc.Query{
		Name:       "hello",
		Parameters: map[string]any{"hello": "world"},
	}

	res, err := node.Query(ctx, conn.LocalAddr().String(), q)
	if err != nil {
		t.Errorf("unexpected error -> %v", err)
	}

	switch res := res.(type) {
	case krpc.Error:
		t.Errorf("error in response -> %v: %v", res.Code(), res.Error())
	case krpc.Response:

	}
}
