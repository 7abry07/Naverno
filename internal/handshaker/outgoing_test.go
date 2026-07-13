package handshaker_test

import (
	"Naverno/internal/handshaker"
	"Naverno/internal/test"
	"bytes"
	"testing"
	"time"
)

func TestValidHandshake(t *testing.T) {
	handshakeSent := []byte{}

	remoteExt := [8]byte{0}
	remoteIh := [20]byte{1}
	remotePid := [20]byte{3}

	handshakeSent = append(handshakeSent, 19)
	handshakeSent = append(handshakeSent, []byte("BitTorrent protocol")...)
	handshakeSent = append(handshakeSent, remoteExt[:]...)
	handshakeSent = append(handshakeSent, remoteIh[:]...)
	handshakeSent = append(handshakeSent, remotePid[:]...)

	conn := test.NewMockConn(handshakeSent)
	result := make(chan handshaker.HandshakeResult)

	ext := [8]byte{0}
	ih := [20]byte{1}
	pid := [20]byte{2}
	ext[0] |= 1 << 7

	outgoing := handshaker.NewOutgoingHandshaker(conn, pid, ih, ext, time.Second*5)
	go outgoing.Run(result)

	handshakedConn := <-result
	if handshakedConn.Error != nil {
		t.Fatalf("unexpected error -> %v", handshakedConn.Error)
	}
	if handshakedConn.PeerID != remotePid {
		t.Errorf("expected peer id -> %#v | got -> %#v", remotePid, handshakedConn.PeerID)
	}
	if handshakedConn.Extensions != remoteExt {
		t.Errorf("expected extension bytes -> %#v | got -> %#v", remoteExt, handshakedConn.Extensions)
	}

	buf := make([]byte, 68)
	_, err := conn.ReadSent(buf)
	if err != nil {
		t.Fatalf("unexpected error -> %v", err)
	}

	if buf[0] != 19 {
		t.Errorf("expected pstrlen -> %#v | got -> %#v", 19, buf[0])
	}
	if !bytes.Equal(buf[1:20], []byte("BitTorrent protocol")) {
		t.Errorf("expected pstr -> %v | got -> %v", "BitTorrent protocol", string(buf[1:20]))
	}
	if !bytes.Equal(buf[20:28], ext[:]) {
		t.Errorf("expected extensions -> %#v | got -> %#v", ext[:], buf[20:28])
	}
	if !bytes.Equal(buf[28:48], ih[:]) {
		t.Errorf("expected info hash -> %#v | got -> %#v", ih, buf[28:48])
	}
	if !bytes.Equal(buf[48:68], pid[:]) {
		t.Errorf("expected peer id -> %#v | got -> %#v", pid, buf[48:68])
	}
}
