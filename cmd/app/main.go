package main

import (
	"Naverno/torrent"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

func main() {
	logger := slog.New(tint.NewHandler(os.Stdout, nil))

	sess := torrent.NewSession(logger)
	_, err := sess.NewTorrentFromFile("internal/metadata/testdata/debian.torrent")
	if err != nil {
		panic(err)
	}
	go func() {
		time.Sleep(time.Second * 5)
		sess.Stop()
	}()

	err = sess.Run()
	if err != nil {
		panic(err)
	}
}
