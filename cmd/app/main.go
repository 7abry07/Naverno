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
	options, err := torrent.FromFile("/home/fabry/Downloads/debian.torrent")
	if err != nil {
		panic(err)
	}
	options.PieceSelectionStrategy = torrent.DEFAULT_PIECE_SELECTION
	options.SavePath = "/home/fabry/Downloads"
	_, err = sess.AddTorrent(options)
	if err != nil {
		panic(err)
	}

	http.ListenAndServe(":6060", nil)
}
