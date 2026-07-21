package main

import (
	"Naverno/torrent"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

func main() {
	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{Level: slog.LevelInfo}))

	sess := torrent.StartSession(logger)
	t, err := sess.NewTorrentFromFile("internal/metadata/testdata/debian.torrent")
	if err != nil {
		panic(err)
	}

	go http.ListenAndServe(":6060", nil)
	time.Sleep(time.Minute * 2)
	t.Stop()

	<-make(chan any)
}
