package choker

import (
	"cmp"
	"slices"
	"sync/atomic"
	"time"
)

type Choker struct {
	unchokeTick    time.Duration
	optimisticTick time.Duration
	downloaders    []Peer
	uninterested   []Peer
	optimistic     Peer
	seeding        atomic.Bool

	interestedC chan Peer

	closeC chan struct{}
	doneC  chan struct{}
}

func New(unchokeTick time.Duration, optimisticTick time.Duration) *Choker {
	return &Choker{
		unchokeTick:    unchokeTick,
		optimisticTick: optimisticTick,
		closeC:         make(chan struct{}),
		doneC:          make(chan struct{}),
	}
}

func (c *Choker) Run(events chan any) {
	unchokeTick := time.NewTicker(c.unchokeTick)
	optimisticTick := time.NewTicker(c.optimisticTick)

	defer close(c.doneC)

	for {
		select {
		case <-c.closeC:
			unchokeTick.Stop()
			optimisticTick.Stop()
			return
		case <-unchokeTick.C:
			events <- Unchoke{}
		case <-optimisticTick.C:
			events <- Optimistic{}
		}
	}
}

func (c *Choker) Downloaders() []Peer {
	return c.downloaders
}

func (c *Choker) Uninterested() []Peer {
	return c.uninterested
}

func (c *Choker) Optimistic() Peer {
	return c.optimistic
}

func (c *Choker) PickPeers(peers []Peer) []Peer {
	peers_ := slices.Clone(peers)
	result := []Peer{}
	interested := []Peer{}
	uninterested := []Peer{}

	for _, p := range peers_ {
		if p.IsInterested() {
			interested = append(interested, p)
		} else {
			uninterested = append(uninterested, p)
		}
	}

	slices.SortFunc(interested, func(e1, e2 Peer) int { return c.compare(e1, e2) })
	slices.SortFunc(uninterested, func(e1, e2 Peer) int { return c.compare(e1, e2) })

	result = append(result, interested[0:min(len(interested), 4)]...)
	c.downloaders = append(c.downloaders, result...)
	for _, p := range uninterested {
		if c.betterThanWorst(uninterested, p) {
			c.uninterested = append(c.uninterested, p)
			result = append(result, p)
		}
	}

	return result
}

func (c *Choker) PickOptimistic(peers []Peer) Peer {
	peers_ := slices.Clone(peers)
	slices.SortFunc(peers_, func(e1, e2 Peer) int { return e1.ConnectedAt().Compare(e2.ConnectedAt()) })
	for _, p := range peers_ {
		if p == c.optimistic ||
			slices.Contains(c.downloaders, p) ||
			slices.Contains(c.uninterested, p) {
			continue
		}
		c.optimistic = p
		return p
	}
	return nil
}

func (c *Choker) OnInterested(p Peer) {
	if slices.Contains(c.uninterested, p) {
		idx := slices.Index(c.uninterested, p)
		c.uninterested = slices.Delete(c.uninterested, idx, idx+1)

		if len(c.downloaders) < 4 {
			c.downloaders = append(c.downloaders, p)
			return
		}

		worst := slices.MinFunc(c.downloaders, func(e1, e2 Peer) int { return c.compare(e1, e2) })
		idx = slices.Index(c.downloaders, worst)
		c.downloaders[idx] = p
	}
}

func (c *Choker) Close() {
	close(c.closeC)
	<-c.doneC
}

func (c *Choker) Seeding() {
	c.seeding.Store(true)
}

func (c *Choker) betterThanWorst(peers []Peer, p Peer) bool {
	if c.seeding.Load() {
		if p.DownloadRate() > slices.MinFunc(peers, func(e1, e2 Peer) int { return c.compare(e1, e2) }).DownloadRate() {
			return true
		}
	} else {
		if p.UploadRate() > slices.MinFunc(peers, func(e1, e2 Peer) int { return c.compare(e1, e2) }).UploadRate() {
			return true
		}
	}
	return false
}

func (c *Choker) compare(e1, e2 Peer) int {
	if c.seeding.Load() {
		return cmp.Compare(e1.DownloadRate(), e2.DownloadRate())
	}
	return cmp.Compare(e1.UploadRate(), e2.UploadRate())
}
