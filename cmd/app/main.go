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
	// logfile, err := os.OpenFile("/home/fabry/Downloads/logfile.txt", os.O_CREATE|os.O_RDWR, 0655)
	// if err != nil {
	// 	panic(err)
	// }
	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{Level: slog.LevelInfo}))

	sess := torrent.StartSession(logger, "/home/fabry/Downloads")
	_, err := sess.NewTorrentFromFile("/home/fabry/Downloads/fedora.torrent")
	if err != nil {
		panic(err)
	}

	go http.ListenAndServe(":6060", nil)
	time.Sleep(time.Minute * 2)

	<-make(chan any)
}
