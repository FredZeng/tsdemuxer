package tsdemuxer

import (
	"context"
	"errors"
	"fmt"
	"io"
)

const SyncByte = 0x47

var (
	ErrNoMorePackets                = errors.New("tsdemuxer: no more packets")
	ErrPacketMustStartWithASyncByte = errors.New("tsdemuxer: packet must start with a sync byte")
)

type Demuxer struct {
	ctx          context.Context
	r            io.Reader
	packetBuffer *packetBuffer
}

func NewDemuxer(ctx context.Context, r io.Reader) *Demuxer {
	return &Demuxer{
		ctx:          ctx,
		r:            r,
		packetBuffer: newPacketBuffer(r, 188),
	}
}

func (t *Demuxer) Demux() (err error) {
	for {
		var p *Packet

		if p, err = t.packetBuffer.next(); err != nil {
			if errors.Is(err, ErrNoMorePackets) {
				err = nil
				return
			}
			err = fmt.Errorf("tsdemuxer: getting next packet from buffer failed: %w", err)
			return
		}

		// TODO:
		fmt.Println(p)
	}
}
