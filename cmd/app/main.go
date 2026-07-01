package main

import (
	"GoBit/internal/metadata"
	"fmt"
	"os"
)

func main() {
	data, err := os.ReadFile("internal/metadata/testdata/debian.torrent")
	if err != nil {
		panic(err)
	}

	meta, err := metadata.Parse(string(data))
	if err != nil {
		panic(err)
	} else {
		fmt.Print(meta)
		return
	}
}
