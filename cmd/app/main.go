package main

import (
	"Naverno/torrent"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/lmittmann/tint"
)

func main() {
	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{Level: slog.LevelInfo}))
	sess := torrent.StartSession(logger)
	// options, err := torrent.FromFile("/home/fabry/Downloads/debian.torrent")
	options, err := torrent.FromURI("magnet:?xt=urn:btih:23c72642e822727c85499efca111fabdf36531bf&dn=%5BToonsHub%5D%20THE%20GHOST%20IN%20THE%20SHELL%20S01E08%201080p%20AMZN%20WEB-DL%20DUAL%20DDP2.0%20H.265%20%28Koukaku%20Kidoutai%3A%20THE%20GHOST%20IN%20THE%20SHELL%2C%20Dual-Audio%2C%20Multi-Subs%29&tr=http%3A%2F%2Fnyaa.tracker.wf%3A7777%2Fannounce&tr=udp%3A%2F%2Fopen.stealth.si%3A80%2Fannounce&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337%2Fannounce&tr=udp%3A%2F%2Fexodus.desync.com%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.torrent.eu.org%3A451%2Fannounce")
	if err != nil {
		panic(err)
	}
	options.PieceSelectionStrategy = torrent.DEFAULT_PIECE_SELECTION
	options.SavePath = "/home/fabry/Downloads"
	t, err := sess.AddTorrent(options)
	if err != nil {
		panic(err)
	}

	t.AnnounceToAllTrackers()

	http.ListenAndServe(":6060", nil)
}
