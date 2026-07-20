package peer

import (
	"Naverno/internal/peer/reader"
	"Naverno/internal/peer/writer"
	"Naverno/internal/peerprotocol"
	"log/slog"
	"net"
	"time"

	"github.com/bits-and-blooms/bitset"
)

type Peer struct {
	logger *slog.Logger

	ID            [20]byte
	Extensions    [8]byte
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

func New(logger *slog.Logger, conn net.Conn, ID [20]byte, extensions [8]byte) *Peer {
	if conn == nil {
		panic("passed nil connection to Peer constructor")
	}
	if logger == nil {
		panic("passed nil logger to Peer constructor")
	}

	return &Peer{
		ID:            ID,
		Extensions:    extensions,
		logger:        logger.With("PeerID", string(ID[:])),
		conn:          conn,
		IsChoked:      true,
		AmChoked:      true,
		IsInteresting: false,
		AmInteresting: false,
		Pieces:        nil,
		out:           writer.New(conn),
		in:            reader.New(conn),
		closeC:        make(chan struct{}),
		doneC:         make(chan struct{}),
	}
}

func (p *Peer) Addr() net.Addr {
	return p.conn.RemoteAddr()
}

func (p *Peer) GetPieces() *bitset.BitSet {
	return p.Pieces
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

	peerTimeout := time.NewTimer(time.Minute * 3)
	selfTimeout := time.NewTicker(time.Minute)

	for {
		select {
		case <-p.closeC:
			return
		case <-selfTimeout.C:
			p.out.Write(peerprotocol.KeepAlive{})
		case <-peerTimeout.C:
			p.logger.Debug("peer -> timeout")
			select {
			case disconnected <- p:
			case <-p.closeC:
			}
			return
		case err := <-p.in.Error():
			p.logger.Debug("peer -> read error", "Error", err.Error())
			select {
			case disconnected <- p:
			case <-p.closeC:
			}
			return
		case err := <-p.out.Error():
			p.logger.Debug("peer -> write error", "error", err.Error())
			select {
			case disconnected <- p:
			case <-p.closeC:
			}
			return
		case mess, ok := <-p.in.Messages():
			if !ok {
				continue
			}
			peerTimeout = time.NewTimer(time.Minute * 3)
			if mess.ID() == peerprotocol.KeepAliveID {
				continue
			}
			inbox <- PeerMessage{p, mess}
		}
	}
}

func (p *Peer) Stop() {
	close(p.closeC)
	p.conn.Close()
	p.out.Close()
	p.in.Close()
	<-p.doneC
}

func (p *Peer) Choke() {
	p.IsChoked = true
	p.out.Write(peerprotocol.Choke{})
}

func (p *Peer) Unchoke() {
	p.IsChoked = false
	p.out.Write(peerprotocol.Unchoke{})
}

func (p *Peer) Interesting() {
	p.IsInteresting = true
	p.out.Write(peerprotocol.Interested{})
}

func (p *Peer) Uninteresting() {
	p.IsInteresting = false
	p.out.Write(peerprotocol.Uninterested{})
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
