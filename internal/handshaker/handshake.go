package handshaker

import (
	"bytes"
	"fmt"
	"io"
	"net"
)

const (
	ProtocolStringLength = byte(19)
	ProtocolString       = "BitTorrent protocol"
)

type HandshakeResult struct {
	Conn       net.Conn
	PeerID     [20]byte
	Extensions [8]byte
	Error      error
}

type Handshake struct {
	Extensions [8]byte
	InfoHash   [20]byte
	PeerID     [20]byte
}

func (h Handshake) Marshal() []byte {
	buf := []byte{}
	buf = append(buf, ProtocolStringLength)
	buf = append(buf, []byte(ProtocolString)...)
	buf = append(buf, h.Extensions[:]...)
	buf = append(buf, h.InfoHash[:]...)
	buf = append(buf, h.PeerID[:]...)
	return buf
}

func (h *Handshake) Unmarshal(in io.Reader) error {
	readBuf := make([]byte, 68)
	_, err := io.ReadFull(in, readBuf)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(readBuf)

	pstrlen, _ := buf.ReadByte()
	pstr := make([]byte, len(ProtocolString))

	buf.Read(pstr)
	buf.Read(h.Extensions[:])
	buf.Read(h.InfoHash[:])
	buf.Read(h.PeerID[:])

	if pstrlen != 19 {
		return fmt.Errorf("protocol string length is invalid")
	}

	if !bytes.Equal(pstr, []byte(ProtocolString)) {
		return fmt.Errorf("protocol string is invalid")
	}

	return nil
}
