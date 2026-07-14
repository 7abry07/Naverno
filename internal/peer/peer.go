package peer

import (
	"Naverno/internal/peer/reader"
	"Naverno/internal/peer/writer"
	"Naverno/internal/peerprotocol"
	"net"
	"time"

	"github.com/bits-and-blooms/bitset"
)

type Peer struct {
	ID            [20]byte
	IsChoked      bool
	IsInteresting bool
	AmChoked      bool
	AmInteresting bool
	Pieces        *bitset.BitSet

	conn net.Conn
	out  *writer.Writer
	in   *reader.Reader

	closeC chan struct{}
	doneC  chan struct{}
}

type PeerMessage struct {
	*Peer
	Message peerprotocol.Message
}

func New(ID [20]byte, conn net.Conn, pieceCount uint32) *Peer {
	if conn == nil {
		panic("passed nil connection to Peer constructor")
	}

	bitset.New(uint(pieceCount))

	return &Peer{
		conn:          conn,
		IsChoked:      true,
		AmChoked:      true,
		IsInteresting: false,
		AmInteresting: false,
		Pieces:        bitset.New(uint(pieceCount)),
		out:           writer.New(conn),
		in:            reader.New(conn),
		closeC:        make(chan struct{}),
		doneC:         make(chan struct{}),
	}
}

func (p *Peer) HasPiece(idx uint32) bool {
	return p.Pieces.Test(uint(idx))
}

func (p *Peer) SetPiece(idx uint32) {
	p.Pieces.Set(uint(idx))
}

func (p *Peer) IsChoking() bool {
	return p.AmChoked
}

func (p *Peer) IsInterested() bool {
	return p.AmInteresting
}

func (p *Peer) Choking() bool {
	return p.IsChoked
}

func (p *Peer) Interested() bool {
	return p.IsInteresting
}

func (p *Peer) Run(inbox chan<- PeerMessage, disconnected chan<- *Peer) {
	defer close(p.doneC)
	go p.in.Run()
	go p.out.Run()

	peerTimeout := time.NewTimer(time.Minute * 2)
	selfTimeout := time.NewTicker(time.Minute)

	for {
		select {
		case <-p.closeC:
			p.conn.Close()
			p.out.Close()
			p.in.Close()
			disconnected <- p
			return
		case <-selfTimeout.C:
			p.out.Write(peerprotocol.KeepAlive{})
		case <-peerTimeout.C:
			close(p.closeC)
		case <-p.in.Error():
			close(p.closeC)
		case <-p.out.Error():
			close(p.closeC)
		case mess := <-p.in.Messages():
			peerTimeout = time.NewTimer(time.Minute * 2)
			inbox <- PeerMessage{p, mess}
		}
	}
}

func (p *Peer) Stop() {
	close(p.closeC)
	<-p.doneC
}

func (p *Peer) Choke() {
	if !p.IsChoked {
		p.out.Write(peerprotocol.Choke{})
		p.IsChoked = true
	}
}

func (p *Peer) Unchoke() {
	if p.IsChoked {
		p.out.Write(peerprotocol.Unchoke{})
		p.IsChoked = false
	}
}

func (p *Peer) Interesting() {
	if !p.IsInteresting {
		p.out.Write(peerprotocol.Interested{})
		p.IsInteresting = true
	}
}

func (p *Peer) Uninteresting() {
	if p.IsInteresting {
		p.out.Write(peerprotocol.Uninterested{})
		p.IsInteresting = false
	}
}

func (p *Peer) Bitfield(pieces []byte) {
	p.out.Write(peerprotocol.Bitfield{Pieces: pieces})
}

func (p *Peer) Have(idx uint32) {
	p.out.Write(peerprotocol.Have{Idx: idx})
}

func (p *Peer) Request(idx uint32, begin uint32, length uint32) {
	p.out.Write(peerprotocol.Request{Idx: idx, Begin: begin, Length: length})
}

func (p *Peer) Piece(idx uint32, begin uint32, data []byte) {
	p.out.Write(peerprotocol.Piece{Idx: idx, Begin: begin, Data: data})
}

func (p *Peer) Cancel(idx uint32, begin uint32, length uint32) {
	p.out.Write(peerprotocol.Cancel{Idx: idx, Begin: begin, Length: length})
}
