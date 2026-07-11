package peer

import (
	"Naverno/internal/peerprotocol"
	"encoding/binary"
	"io"
	"net"
	"time"
)

const (
	selfTimeoutDuration = time.Minute * 2
	peerTimeoutDuration = time.Minute * 2
)

type Peer struct {
	ID            [20]byte
	IsChoked      bool
	IsInteresting bool
	AmChoked      bool
	AmInteresting bool

	conn net.Conn

	Pieces           []byte
	bitfieldReceived bool

	selfTimeout *time.Ticker
	peerTimeout *time.Timer

	out chan peerprotocol.Message
	in  chan peerprotocol.Message

	closeC chan struct{}
	doneC  chan struct{}
}

func New(ID [20]byte, conn net.Conn) *Peer {
	if conn == nil {
		panic("passed nil connection to Peer constructor")
	}

	return &Peer{
		conn:             conn,
		IsChoked:         true,
		AmChoked:         true,
		IsInteresting:    false,
		AmInteresting:    false,
		bitfieldReceived: false,
		Pieces:           []byte{},
		selfTimeout:      time.NewTicker(selfTimeoutDuration),
		peerTimeout:      time.NewTimer(peerTimeoutDuration),
		out:              make(chan peerprotocol.Message),
		in:               make(chan peerprotocol.Message),
		closeC:           make(chan struct{}),
		doneC:            make(chan struct{}),
	}
}

type PeerMessage struct {
	*Peer
	Message peerprotocol.Message
}

func (p *Peer) Run(inbox chan<- PeerMessage, disconnected chan<- *Peer) {
	for {
		select {
		case <-p.closeC:
			p.conn.Close()
			disconnected <- p
			close(p.doneC)
			return
		case <-p.peerTimeout.C:
			close(p.closeC)
		case <-p.selfTimeout.C:
			p.writeMessage(peerprotocol.KeepAlive{})
		case mess := <-p.in:
			p.peerTimeout = time.NewTimer(peerTimeoutDuration)

			if mess.ID() == peerprotocol.BitfieldID && p.bitfieldReceived {
				if p.bitfieldReceived {
					close(p.closeC)
					break
				}
				p.bitfieldReceived = true
			}

			inbox <- PeerMessage{p, mess}
		case mess := <-p.out:
			p.writeMessage(mess)
		}
	}
}

func (p *Peer) Stop() <-chan struct{} {
	close(p.closeC)
	return p.doneC
}

func (p *Peer) listen() {
	lengthBytes := make([]byte, 4)
	_, err := io.ReadFull(p.conn, lengthBytes)
	if err != nil {
		close(p.closeC)
	}
	length := binary.BigEndian.Uint32(lengthBytes)

	messBytes := make([]byte, length)
	_, err = io.ReadFull(p.conn, messBytes)

	fullMess := []byte{}
	fullMess = append(fullMess, lengthBytes...)
	fullMess = append(fullMess, messBytes...)

	mess, err := peerprotocol.Decode(fullMess)
	if err != nil {
		close(p.closeC)
	}

	p.in <- mess
}

func (p *Peer) writeMessage(mess peerprotocol.Message) error {
	data := mess.Marshal()
	for len(data) > 0 {
		n, err := p.conn.Write(data)
		if err != nil {
			return err
		}
		data = data[n:]
	}
	return nil
}
