package main

import (
	"Naverno/torrent"
	"github.com/lmittmann/tint"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
)

func main() {
	logger := slog.New(tint.NewHandler(os.Stdout, nil))

	sess := torrent.StartSession(logger)
	_, err := sess.NewTorrentFromFile("internal/metadata/testdata/debian.torrent")
	if err != nil {
		panic(err)
	}

	go http.ListenAndServe(":6060", nil)

	c := make(chan any)
	<-c
}
