package main

import (
	"context"
	"os"
	"tsdemuxer"
)

func main() {
	ctx, _ := context.WithCancel(context.Background())

	f, err := os.Open("./test.ts")

	if err != nil {
		panic(err)
		return
	}

	defer f.Close()

	demuxer := tsdemuxer.NewDemuxer(ctx, f)

	if err := demuxer.Demux(); err != nil {
		panic(err)
	}
}
