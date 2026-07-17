package announcer

import (
	"Naverno/internal/tracker"
	"context"
	"log/slog"
	"net/netip"
	"time"
)

type Announcer struct {
	trackers []tracker.Tracker
	logger   *slog.Logger
	numwant  uint32
	port     uint16

	torrentC chan Torrent

	closeC chan struct{}
	doneC  chan struct{}
}

func NewAnnouncer(logger *slog.Logger, torrentC chan Torrent, trackers []tracker.Tracker, port uint16) *Announcer {
	if logger == nil {
		panic("passed nil logger to Announcer")
	}

	a := Announcer{
		trackers: trackers,
		logger:   logger,
		torrentC: torrentC,
		numwant:  200,
		port:     port,

		closeC: make(chan struct{}),
		doneC:  make(chan struct{}),
	}

	return &a
}

func (a *Announcer) Run(peers chan []netip.AddrPort) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	defer close(a.doneC)

	announceTimer := time.NewTimer(0)

	for _, tr := range a.trackers {
		res, err := a.announce(ctx, tr, tracker.TRACKER_STARTED)
		if err != nil {
			a.logger.Warn("announcer -> error in tracker response", "Tracker URL", tr.URL(), "Error", err)
			continue
		}
		announceTimer = time.NewTimer(res.Interval)
		a.logger.Warn("announcer -> announced succesfully", "Tracker URL", tr.URL(), "Reannounce In", res.Interval.Seconds())
		peers <- res.Peers
	}

	for {
		select {
		case <-a.closeC:
			{
				for _, tr := range a.trackers {
					go a.announce(ctx, tr, tracker.TRACKER_STOPPED)
				}
				return
			}
		case <-announceTimer.C:
			{
				for _, tr := range a.trackers {
					res, err := a.announce(ctx, tr, tracker.TRACKER_NONE)
					if err != nil {
						a.logger.Warn("announcer -> error in tracker response", "Tracker URL", tr.URL(), "Error", err)
						continue
					}
					announceTimer = time.NewTimer(res.Interval)
					peers <- res.Peers
					break
				}
			}
		}
	}
}

func (a *Announcer) announce(ctx context.Context, tr tracker.Tracker, event tracker.TrackerEvent) (*tracker.AnnounceResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	if event != tracker.TRACKER_STOPPED {
		a.torrentC <- Torrent{}
	}
	torrent := <-a.torrentC

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

	return tr.Announce(ctx, req)
}

func (a *Announcer) Close(torrent Torrent) {
	close(a.closeC)
	a.torrentC <- torrent
	<-a.doneC
}
