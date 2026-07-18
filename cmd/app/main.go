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
	logger := slog.New(tint.NewHandler(os.Stdout, nil))

	sess := torrent.StartSession(logger)
	_, err := sess.NewTorrentFromFile("internal/metadata/testdata/debian.torrent")
	if err != nil {
		panic(err)
	}

	go http.ListenAndServe(":6060", nil)
	time.Sleep(time.Second * 10)

	c := make(chan any)
	<-c
}
