package peer

import (
	"Naverno/internal/peerprotocol"
	"net"
)

type Peer struct {
	ID            [20]byte
	IsChoked      bool
	IsInteresting bool
	AmChoked      bool
	AmInteresting bool

	conn net.Conn

	pieces []byte

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
		conn:          conn,
		IsChoked:      true,
		AmChoked:      true,
		IsInteresting: false,
		AmInteresting: false,
		pieces:        []byte{},
		out:           make(chan peerprotocol.Message),
		in:            make(chan peerprotocol.Message),
		closeC:        make(chan struct{}),
		doneC:         make(chan struct{}),
	}
}

type PeerMessage struct {
	*Peer
	Message peerprotocol.Message
}

func (p *Peer) Run(inbox chan<- PeerMessage) {
	select {
	case <-p.closeC:
		p.conn.Close()
		close(p.doneC)
	case mess := <-p.in:
		inbox <- PeerMessage{p, mess}
	case _ = <-p.out:
		// TODO
	}
}

func (p *Peer) Choke() {
	if p.IsChoked {
		return
	}
	p.out <- peerprotocol.Choke{}
}

func (p *Peer) Unchoke() {
	if !p.IsChoked {
		return
	}
	p.out <- peerprotocol.Unchoke{}
}

func (p *Peer) Interesting() {
	if p.IsInteresting {
		return
	}
	p.out <- peerprotocol.Interested{}
}

func (p *Peer) Uninteresting() {
	if !p.IsInteresting {
		return
	}
	p.out <- peerprotocol.Uninterested{}
}

func (p *Peer) Stop() <-chan struct{} {
	close(p.closeC)
	return p.doneC
}

func (p *Peer) listen() {
	// TODO
}
