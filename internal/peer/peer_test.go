package peer_test

import (
	"Naverno/internal/peer"
	"Naverno/internal/peerprotocol"
	"bytes"
	"net"
	"reflect"
	"testing"
	"time"
)

type MockConn struct {
	readBuf  bytes.Buffer
	WriteBuf bytes.Buffer
}

func NewMockConn(data []byte) *MockConn {
	return &MockConn{
		readBuf:  *bytes.NewBuffer(data),
		WriteBuf: *bytes.NewBuffer([]byte{}),
	}
}

func (c *MockConn) Read(b []byte) (int, error) {
	return c.readBuf.Read(b)
}

func (c *MockConn) Write(b []byte) (int, error) {
	return c.WriteBuf.Write(b)
}

func (c *MockConn) Close() error                       { return nil }
func (c *MockConn) LocalAddr() net.Addr                { return nil }
func (c *MockConn) RemoteAddr() net.Addr               { return nil }
func (c *MockConn) SetDeadline(t time.Time) error      { return nil }
func (c *MockConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *MockConn) SetWriteDeadline(t time.Time) error { return nil }

func TestPeerMessages(t *testing.T) {
	incomingMessC := make(chan peer.PeerMessage)
	disconnectingC := make(chan *peer.Peer)

	messagesExp := []peerprotocol.Message{}
	messagesRec := []peerprotocol.Message{}

	messagesExp = append(messagesExp, peerprotocol.Bitfield{Pieces: make([]byte, 10)})
	messagesExp = append(messagesExp, peerprotocol.KeepAlive{})
	messagesExp = append(messagesExp, peerprotocol.Choke{})
	messagesExp = append(messagesExp, peerprotocol.Unchoke{})
	messagesExp = append(messagesExp, peerprotocol.Interested{})
	messagesExp = append(messagesExp, peerprotocol.Uninterested{})
	messagesExp = append(messagesExp, peerprotocol.Have{Idx: 5})
	messagesExp = append(messagesExp, peerprotocol.Request{Idx: 5, Begin: 500, Length: 100})
	messagesExp = append(messagesExp, peerprotocol.Piece{Idx: 5, Begin: 500, Data: make([]byte, 100)})
	messagesExp = append(messagesExp, peerprotocol.Cancel{Idx: 5, Begin: 500, Length: 100})

	buf := []byte{}
	for _, m := range messagesExp {
		buf = append(buf, m.Marshal()...)
	}
	conn := NewMockConn(buf)

	p := peer.New([20]byte{}, conn, 80)
	go p.Run(incomingMessC, disconnectingC)

	for range len(messagesExp) {
		select {
		case p := <-incomingMessC:
			messagesRec = append(messagesRec, p.Message)
		case _ = <-disconnectingC:
			t.Fatal("peer disconnected")
		}
	}

	if !reflect.DeepEqual(messagesExp, messagesRec) {
		t.Fatal("messages read by peer are not equal to the messages that were actually sent")
	}
}
