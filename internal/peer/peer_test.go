package peer_test

import (
	"Naverno/internal/peer"
	"Naverno/internal/peerprotocol"
	"Naverno/internal/piece"
	"Naverno/internal/test"
	"io"
	"log/slog"
	"reflect"
	"testing"
	"time"
)

func TestRead(t *testing.T) {
	incomingMessC := make(chan peer.PeerMessage)
	disconnectingC := make(chan *peer.Peer)

	messagesExp := []peerprotocol.Message{}
	messagesRec := []peerprotocol.Message{}

	messagesExp = append(messagesExp, peerprotocol.Choke{})
	messagesExp = append(messagesExp, peerprotocol.Unchoke{})
	messagesExp = append(messagesExp, peerprotocol.Interested{})
	messagesExp = append(messagesExp, peerprotocol.Uninterested{})
	messagesExp = append(messagesExp, peerprotocol.Bitfield{Pieces: make([]byte, 10)})
	messagesExp = append(messagesExp, peerprotocol.Have{Idx: 5})
	messagesExp = append(messagesExp, peerprotocol.Request{Idx: 5, Begin: 500, Length: 100})
	messagesExp = append(messagesExp, peerprotocol.Piece{Idx: 5, Begin: 500, Data: make([]byte, 100)})
	messagesExp = append(messagesExp, peerprotocol.Cancel{Idx: 5, Begin: 500, Length: 100})

	buf := []byte{}
	for _, m := range messagesExp {
		buf = append(buf, m.Marshal()...)
	}
	conn := test.NewMockConn(buf)

	p := peer.New(slog.New(slog.NewTextHandler(io.Discard, nil)), conn, [20]byte{}, [8]byte{})

	go p.Run(incomingMessC, disconnectingC)

	testTime := time.NewTimer(time.Second * 5)
	for range len(messagesExp) {
		select {
		case p := <-incomingMessC:
			messagesRec = append(messagesRec, p.Message)
		case <-disconnectingC:
			t.Fatal("peer disconnected")
		case <-testTime.C:
			t.Fatal("test time was excedded")
		}
	}

	if !reflect.DeepEqual(messagesExp, messagesRec) {
		t.Fatal("messages read by peer are not equal to the messages that were actually sent")
	}
}

func TestWrite(t *testing.T) {
	incomingMessC := make(chan peer.PeerMessage)
	disconnectingC := make(chan *peer.Peer)

	conn := test.NewMockConn([]byte{})
	p := peer.New(slog.New(slog.NewTextHandler(io.Discard, nil)), conn, [20]byte{}, [8]byte{})
	go p.Run(incomingMessC, disconnectingC)

	ok := p.Have(piece.NewPiece(5, 0, 0, [20]byte{}))
	if !ok {
		t.Errorf("sending message failed")
	}
}

func TestInvalidRead(t *testing.T) {
	incomingMessC := make(chan peer.PeerMessage)
	disconnectingC := make(chan *peer.Peer)

	buf := []byte{}
	buf = append(buf, []byte{0, 0, 0, 1, 255}...)
	conn := test.NewMockConn(buf)

	p := peer.New(slog.New(slog.NewTextHandler(io.Discard, nil)), conn, [20]byte{}, [8]byte{})
	go p.Run(incomingMessC, disconnectingC)

	testTime := time.NewTimer(time.Second * 5)
	disconnected := false
	select {
	case <-incomingMessC:
	case <-disconnectingC:
		disconnected = true
	case <-testTime.C:
		t.Fatal("test time was excedded")
	}

	if !disconnected {
		t.Fatal("peer should have disconnected but didn't")
	}
}
