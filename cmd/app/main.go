package main

import (
	"Naverno/torrent"
	"context"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

func main() {
	logger := slog.New(tint.NewHandler(os.Stdout, nil))

	sess := torrent.StartSession(context.Background(), logger)
	_, err := sess.NewTorrentFromFile("internal/metadata/testdata/debian.torrent")
	if err != nil {
		panic(err)
	}

	c := make(chan any)
	<-c
}
