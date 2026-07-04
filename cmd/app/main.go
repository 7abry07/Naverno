package main

import (
	"Naverno/internal/metadata"
	"Naverno/internal/tracker"
	"Naverno/internal/trackermanager"

	"context"
	"fmt"
	"github.com/lmittmann/tint"
	"log/slog"
	"os"
)

func main() {
	fmt.Println("hello world")

	file, err := os.Open("internal/metadata/testdata/fedora.torrent")
	if err != nil {
		panic(err)
	}

	meta, err := metadata.New(file)
	if err != nil {
		panic(err)
	}

	for _, file := range meta.Files {
		fmt.Printf("file -> %v\n", file)
	}

	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{
		Level: slog.LevelDebug,
	}))

	manager := trackermanager.New(logger)

	t, err := manager.Get("http://torrent.fedoraproject.org:6969/announce")
	if err != nil {
		panic(err)
	}

	pid := [20]byte{}
	copy(pid[:], "-GB0001-hfjdjakfjfld")

	req := tracker.AnnounceRequest{
		Infohash:   meta.Infohash,
		PeerID:     pid,
		Downloaded: 0,
		Uploaded:   0,
		Left:       0,
		Port:       6881,
		Numwant:    200,
		Event:      tracker.TRACKER_STARTED,
	}

	resp, err := t.Announce(context.Background(), req)
	if err != nil {
		panic(err)
	}

	for _, p := range resp.Peers {
		fmt.Printf("%v:%v\n", p.Addr(), p.Port())
	}
}
