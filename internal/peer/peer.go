package peer

import (
	"Naverno/internal/bitfield"
	"Naverno/internal/peer/reader"
	"Naverno/internal/peer/writer"
	"Naverno/internal/peerprotocol"
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

	connectedAt     time.Time
	lastRateUpdated time.Time
	Downloaded      uint64
	Uploaded        uint64
	downloadRate    uint64
	uploadRate      uint64

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

func New(logger *slog.Logger, connectedAt time.Time, conn net.Conn, ID [20]byte, extensions [8]byte) *Peer {
	if conn == nil {
		panic("passed nil connection to Peer constructor")
	}
	if logger == nil {
		panic("passed nil logger to Peer constructor")
	}

	plogger := logger.With("PeerID", string(ID[:]))
	return &Peer{
		ID:              ID,
		connectedAt:     connectedAt,
		Extensions:      extensions,
		logger:          plogger,
		conn:            conn,
		IsChoked:        true,
		AmChoked:        true,
		IsInteresting:   false,
		AmInteresting:   false,
		lastRateUpdated: time.Now(),
		Pieces:          nil,
		out:             writer.New(plogger, conn),
		in:              reader.New(plogger, conn),
		closeC:          make(chan struct{}),
		doneC:           make(chan struct{}),
	}
}

func (p *Peer) Addr() net.Addr {
	return p.conn.RemoteAddr()
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

func (p *Peer) ConnectedAt() time.Time {
	return p.connectedAt
}

func (p *Peer) DownloadRate() uint64 {
	return p.downloadRate
}

func (p *Peer) UploadRate() uint64 {
	return p.uploadRate
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

func (p *Peer) CalculateStats(t time.Time) {
	p.uploadRate = p.Uploaded / uint64(time.Since(p.lastRateUpdated).Milliseconds())
	p.downloadRate = p.Downloaded / uint64(time.Since(p.lastRateUpdated).Milliseconds())
	p.lastRateUpdated = t
	p.Downloaded = 0
	p.Uploaded = 0
}

func (p *Peer) Stop() {
	close(p.closeC)
	p.conn.Close()
	p.out.Close()
	p.in.Close()
	<-p.doneC
}

func (p *Peer) Choke() {
	if p.IsChoked {
		return
	}
	p.out.Write(peerprotocol.Choke{})
	p.IsChoked = true
}

func (p *Peer) Unchoke() {
	if !p.IsChoked {
		return
	}
	p.out.Write(peerprotocol.Unchoke{})
	p.IsChoked = false
}

func (p *Peer) Interesting() {
	if p.IsInteresting {
		return
	}
	p.out.Write(peerprotocol.Interested{})
	p.IsInteresting = true
}

func (p *Peer) Uninteresting() {
	if !p.IsInteresting {
		return
	}
	p.out.Write(peerprotocol.Uninterested{})
	p.IsInteresting = false
}

func (p *Peer) Bitfield(pieces []byte) {
	p.out.Write(peerprotocol.Bitfield{Pieces: pieces})
}

func (p *Peer) Have(idx uint32) {
	p.out.Write(peerprotocol.Have{Idx: idx})
}

func (p *Peer) Request(idx, begin uint32, length uint32) {
	p.out.Write(peerprotocol.Request{Idx: idx, Begin: begin, Length: length})
}

func (p *Peer) Piece(idx, begin uint32, data []byte) {
	p.out.Write(peerprotocol.Piece{Idx: idx, Begin: begin, Data: data})
}

func (p *Peer) Cancel(idx, begin, length uint32) {
	p.out.Write(peerprotocol.Cancel{Idx: idx, Begin: begin, Length: length})
}
