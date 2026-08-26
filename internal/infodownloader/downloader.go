package infodownloader

import (
	"Naverno/internal/util"
	"errors"
)

var (
	ErrInvalid      = errors.New("block invalid")
	ErrNotRequested = errors.New("block not requested")
	ErrDuplicate    = errors.New("block duplicate")
)

const (
	PieceSize = 16 * 1024
)

type InfoDownloader struct {
	Size          int
	buffer        [][]byte
	peers         []Peer
	pieces        int
	lastPieceSize uint32
	pending       map[uint32]Peer
	done          map[uint32]struct{}
	remaining     []uint32
}

func New(size int) *InfoDownloader {
	downloader := &InfoDownloader{
		pending: make(map[uint32]Peer),
		done:    make(map[uint32]struct{}),
	}
	downloader.Size = size
	if size <= PieceSize {
		downloader.pieces = 1
		downloader.lastPieceSize = uint32(size)
		downloader.remaining = append(downloader.remaining, 0)
		downloader.buffer = append(downloader.buffer, []byte{})
	} else {
		downloader.pieces = int(util.Align(uint64(size), uint64(PieceSize)) / PieceSize)
		downloader.lastPieceSize = uint32(util.Align(uint64(size), uint64(PieceSize)) - uint64(size))
		for p := range downloader.pieces {
			downloader.remaining = append(downloader.remaining, uint32(p))
		}
		if downloader.lastPieceSize == 0 {
			downloader.lastPieceSize = PieceSize
		}
		for range downloader.pieces {
			downloader.buffer = append(downloader.buffer, []byte{})
		}
	}
	return downloader
}

func (d *InfoDownloader) Request(maxQueue int) {
	remaining := d.remaining
	for _, piece := range remaining {
		if len(d.pending) >= maxQueue {
			return
		}

		peer := d.peers[0]
		d.peers = d.peers[1:]

		d.pending[piece] = peer
		peer.RequestMetadata(piece)
		d.remaining = d.remaining[1:]

		d.peers = append(d.peers, peer)
	}
}

func (d *InfoDownloader) Completed() ([]byte, bool) {
	if len(d.done) == int(d.pieces) {
		buf := []byte{}
		for _, piece := range d.buffer {
			buf = append(buf, piece...)
		}
		return buf, true
	}
	return []byte{}, false
}

func (d *InfoDownloader) OnPiece(piece uint32, data []byte) error {
	if (piece >= uint32(d.pieces)) ||
		(len(data) > PieceSize) ||
		(len(data) < PieceSize && piece != uint32(d.pieces)-1) ||
		(len(data) != int(d.lastPieceSize)) {
		return ErrInvalid
	}

	_, ok := d.pending[piece]
	if !ok {
		return ErrNotRequested
	}
	if _, ok := d.done[piece]; ok {
		return ErrDuplicate
	}
	delete(d.pending, piece)
	d.done[piece] = struct{}{}
	d.buffer[piece] = data
	return nil
}

func (d *InfoDownloader) OnReject(piece uint32) error {
	if piece >= uint32(d.pieces) {
		return ErrInvalid
	}
	_, ok := d.pending[piece]
	if !ok {
		return ErrNotRequested
	}
	d.remaining = append(d.remaining, piece)
	delete(d.pending, piece)
	return nil
}

func (d *InfoDownloader) AddPeer(peer Peer) {
	d.peers = append(d.peers, peer)
}

func (d *InfoDownloader) RemovePeer(peer Peer) {
	for piece, p := range d.pending {
		if p == peer {
			delete(d.pending, piece)
			d.remaining = append(d.remaining, piece)
		}
	}
	temp := []Peer{}
	for _, p := range d.peers {
		if p != peer {
			temp = append(temp, p)
		}
	}
	d.peers = temp
}

func (d *InfoDownloader) Reset() {
	d.remaining = []uint32{}
	d.pending = make(map[uint32]Peer)
	d.done = make(map[uint32]struct{})
	for p := range d.pieces {
		d.buffer[p] = []byte{}
		d.remaining = append(d.remaining, uint32(p))
	}
}
