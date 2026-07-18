package announcer

import (
	"Naverno/internal/tracker"
	"context"
	"log/slog"
	"net/netip"
	"time"
)

type Announcer struct {
	trackers      []tracker.Tracker
	logger        *slog.Logger
	announceTimer *time.Timer
	numwant       uint32
	port          uint16

	closeC chan struct{}
	doneC  chan struct{}
}

func New(logger *slog.Logger, trackers []tracker.Tracker, port uint16) *Announcer {
	if logger == nil {
		panic("passed nil logger to Announcer")
	}

	a := Announcer{
		trackers:      trackers,
		logger:        logger,
		announceTimer: nil,
		numwant:       200,
		port:          port,

		closeC: make(chan struct{}),
		doneC:  make(chan struct{}),
	}

	return &a
}

func (a *Announcer) Run(torrentC chan Torrent, peers chan []netip.AddrPort) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	defer close(a.doneC)

	a.announceTimer = time.NewTimer(0)

	torrentC <- Torrent{}
	torrent := <-torrentC

	for _, tr := range a.trackers {
		res, err := a.announce(ctx, tr, torrent, tracker.TRACKER_STARTED)
		if err != nil {
			a.logger.Warn("announcer -> error in tracker response", "Tracker URL", tr.URL(), "Error", err)
			continue
		}
		a.announceTimer = time.NewTimer(res.Interval)
		a.logger.Warn("announcer -> announced succesfully", "Tracker URL", tr.URL(), "Reannounce In", res.Interval.Seconds())
		peers <- res.Peers
	}

	for {
		select {
		case <-a.closeC:
			torrent := <-torrentC
			for _, tr := range a.trackers {
				go a.announce(ctx, tr, torrent, tracker.TRACKER_STOPPED)
			}
			return
		case <-a.announceTimer.C:
			torrentC <- Torrent{}
			torrent := <-torrentC

			for _, tr := range a.trackers {
				res, err := a.announce(ctx, tr, torrent, tracker.TRACKER_NONE)
				if err != nil {
					a.logger.Warn("announcer -> error in tracker response", "Tracker URL", tr.URL(), "Error", err)
					continue
				}
				a.announceTimer = time.NewTimer(res.Interval)
				peers <- res.Peers
				break
			}
		}
	}
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
		Ip:         netip.Addr{},
		Port:       a.port,
		Numwant:    a.numwant,
		Event:      event,
	}

	if event == tracker.TRACKER_STOPPED {
		req.Numwant = 0
	}

	return tr.Announce(ctx, req)
}

func (a *Announcer) Close() {
	close(a.closeC)
	a.announceTimer.Stop()
	<-a.announceTimer.C
	<-a.doneC
}
