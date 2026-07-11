package peer

import "Naverno/internal/peerprotocol"

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

func (p *Peer) Bitfield(pieces []byte) {
	p.out <- peerprotocol.Bitfield{Pieces: pieces}
}

func (p *Peer) Have(piece uint32) {
	p.out <- peerprotocol.Have{Idx: piece}
}
