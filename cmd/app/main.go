package main

import (
	"Naverno/torrent"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	// "os"
	"time"

	"github.com/lmittmann/tint"
)

func main() {
	logger := slog.New(tint.NewHandler(io.Discard, &tint.Options{Level: slog.LevelInfo}))
	sess := torrent.StartSession(logger)
	t, err := sess.AddTorrentFromFile("/home/fabry/Downloads/debian.torrent", "/home/fabry/Downloads")
	if err != nil {
		panic(err)
	}

	go http.ListenAndServe(":6060", nil)

	ticker := time.NewTicker(time.Second * 1)
	for {
		<-ticker.C
		stats := t.GetStats()
		fmt.Printf("downloaded -> %v\n", stats.Downloaded)
		fmt.Printf("uploaded-> %v\n", stats.Uploaded)
		fmt.Printf("connections -> %v\n", stats.Connections)
		fmt.Printf("peers -> %v\n", len(stats.Peers))
		fmt.Printf("trackers-> %v\n", len(stats.Trackers))
	}

	// <-make(chan any)
}
