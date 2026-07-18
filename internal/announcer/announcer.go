package announcer

import (
	"Naverno/internal/tracker"
	"Naverno/internal/util"
	"context"
	"log/slog"
	"math/rand/v2"
	"net/netip"
	"slices"
	"time"
)

type Announcer struct {
	firstAnnounce map[tracker.Tracker]struct{}
	trackers      [][]tracker.Tracker
	logger        *slog.Logger
	announceTimer *time.Timer
	numwant       uint32
	port          uint16

	closeC chan struct{}
	doneC  chan struct{}
}

func New(logger *slog.Logger, trackers [][]tracker.Tracker, port uint16) *Announcer {
	if logger == nil {
		panic("passed nil logger to Announcer")
	}

	a := Announcer{
		firstAnnounce: make(map[tracker.Tracker]struct{}),
		trackers:      trackers,
		logger:        logger,
		announceTimer: nil,
		numwant:       200,
		port:          port,

		closeC: make(chan struct{}),
		doneC:  make(chan struct{}),
	}

	for _, tier := range trackers {
		rand.Shuffle(len(tier), func(i, j int) { tier[i], tier[j] = tier[j], tier[i] })
		for _, tr := range tier {
			a.firstAnnounce[tr] = struct{}{}
		}
	}

	return &a
}

func (a *Announcer) Run(torrentC chan Torrent, peers chan []netip.AddrPort) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	defer close(a.doneC)

	a.announceTimer = time.NewTimer(0)

	for {
		select {
		case <-a.closeC:
			torrent := <-torrentC
			for _, tier := range a.trackers {
				for _, tr := range tier {
					a.announce(ctx, tr, torrent, tracker.TRACKER_STOPPED)
				}
			}
			return
		case <-a.announceTimer.C:
			torrentC <- Torrent{}
			torrent := <-torrentC

			for i, tier := range a.trackers {
				res, ok := a.announceTier(ctx, tier, torrent)
				if ok {
					a.trackers[i] = tier
					peers <- res
					break
				}
			}
		}
	}
}

func (a *Announcer) announceTier(ctx context.Context, tier []tracker.Tracker, torrent Torrent) ([]netip.AddrPort, bool) {
	for _, tr := range tier {
		_, first := a.firstAnnounce[tr]
		ev := tracker.TRACKER_NONE
		if first {
			ev = tracker.TRACKER_STARTED
		}
		delete(a.firstAnnounce, tr)

		res, err := a.announce(ctx, tr, torrent, ev)
		if err != nil {
			a.logger.Warn("announcer -> error in tracker response", "Tracker URL", tr.URL(), "Error", err)
			continue
		}
		a.announceTimer = time.NewTimer(res.Interval)
		a.logger.Info("announcer -> announced succesfully", "Tracker URL", tr.URL(), "Reannounce In", res.Interval.Seconds())

		tier = util.Remove(tier, tr, func(e1, e2 tracker.Tracker) bool { return e1 == e2 })
		tier = slices.Insert(tier, 0, tr)

		return res.Peers, true
	}
	return []netip.AddrPort{}, false
}

func (a *Announcer) announce(ctx context.Context, tr tracker.Tracker, torrent Torrent, event tracker.TrackerEvent) (*tracker.AnnounceResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	req := tracker.AnnounceRequest{
		Infohash:   torrent.InfoHash,
		PeerID:     torrent.PeerID,
		Downloaded: torrent.Downloaded,
		Uploaded:   torrent.Uploaded,
		Left:       torrent.Left,
		Port:       a.port,
		Numwant:    a.numwant,
		Event:      event,
	}

	if event == tracker.TRACKER_STOPPED {
		req.Numwant = 0
	}

	return tr.Announce(ctx, req)
}

func (a *Announcer) Close(t Torrent, tC chan Torrent) {
	close(a.closeC)
	if !a.announceTimer.Stop() {
		select {
		case <-a.announceTimer.C:
		default:
		}
	}
	tC <- t
	<-a.doneC
}
