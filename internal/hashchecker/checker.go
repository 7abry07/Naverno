package hashchecker

import (
	"Naverno/internal/piece"
	"Naverno/internal/storage"
	"crypto/sha1"
)

type HashChecker struct {
	storage storage.Storage
	Piece   *piece.Piece
	Matches bool
	Err     error

	closeC chan struct{}
	doneC  chan struct{}
}

func New(s storage.Storage, p *piece.Piece) *HashChecker {
	return &HashChecker{
		storage: s,
		Piece:   p,
		Matches: false,
		Err:     nil,
	}
}

func (c *HashChecker) Run(result chan<- *HashChecker) {
	defer close(c.doneC)
	defer func() {
		select {
		case <-c.closeC:
		case result <- c:
		}
	}()

	data, err := c.storage.Read(c.Piece.Offset, c.Piece.Size)
	if err != nil {
		c.Err = err
		return
	}

	if sha1.Sum(data) == c.Piece.Hash {
		c.Matches = true
		return
	}
}

func (c *HashChecker) Close() {
	close(c.closeC)
	<-c.doneC
}
