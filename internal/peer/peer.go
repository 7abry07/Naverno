package peer

import (
	"Naverno/internal/bitfield"
	"Naverno/internal/peer/reader"
	"Naverno/internal/peer/writer"
	"Naverno/internal/peerprotocol"
	"Naverno/internal/piece"
	"log/slog"
	"net"
	"time"
)

type Peer struct {
	logger *slog.Logger

	ID            [20]byte
	Extensions    [8]byte
	IsChoked      bool
	IsInteresting bool
	AmChoked      bool
	AmInteresting bool
	Pieces        *bitfield.Bitfield

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

	plogger := logger.With("PeerID", string(ID[:]))
	return &Peer{
		ID:            ID,
		Extensions:    extensions,
		logger:        plogger,
		conn:          conn,
		IsChoked:      true,
		AmChoked:      true,
		IsInteresting: false,
		AmInteresting: false,
		Pieces:        nil,
		out:           writer.New(plogger, conn),
		in:            reader.New(plogger, conn),
		closeC:        make(chan struct{}),
		doneC:         make(chan struct{}),
	}
}

func (p *Peer) Addr() net.Addr {
	return p.conn.RemoteAddr()
}

func (p *Peer) GetPieces() *bitfield.Bitfield {
	return p.Pieces
}

func (p *Peer) HasPiece(idx uint32) bool {
	return p.Pieces.Test(idx)
}

func (p *Peer) SetPiece(idx uint32) {
	p.Pieces.Set(idx)
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

func (p *Peer) Choke() bool {
	ok := p.out.Write(peerprotocol.Choke{})
	if ok {
		p.IsChoked = true
	}
	return ok
}

func (p *Peer) Unchoke() bool {
	ok := p.out.Write(peerprotocol.Unchoke{})
	if ok {
		p.IsChoked = false
	}
	return ok
}

func (p *Peer) Interesting() bool {
	ok := p.out.Write(peerprotocol.Interested{})
	if ok {
		p.IsInteresting = true
	}
	return ok
}

func (p *Peer) Uninteresting() bool {
	ok := p.out.Write(peerprotocol.Uninterested{})
	if ok {
		p.IsInteresting = false
	}
	return ok
}

func (p *Peer) Bitfield(pieces []byte) bool {
	return p.out.Write(peerprotocol.Bitfield{Pieces: pieces})
}

func (p *Peer) Have(pi *piece.Piece) bool {
	return p.out.Write(peerprotocol.Have{Idx: pi.Idx})
}

func (p *Peer) Request(pi *piece.Piece, begin uint32, length uint32) bool {
	return p.out.Write(peerprotocol.Request{Idx: pi.Idx, Begin: begin, Length: length})
}

func (p *Peer) Piece(pi *piece.Piece, begin uint32, data []byte) bool {
	return p.out.Write(peerprotocol.Piece{Idx: pi.Idx, Begin: begin, Data: data})
}

func (p *Peer) Cancel(pi *piece.Piece, begin uint32, length uint32) bool {
	return p.out.Write(peerprotocol.Cancel{Idx: pi.Idx, Begin: begin, Length: length})
}
