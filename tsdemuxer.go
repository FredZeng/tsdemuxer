package tsdemuxer

import (
	"context"
	"errors"
	"io"
)

const SyncByte = 0x47

var (
	ErrPacketTooSmall               = errors.New("tsdemuxer: packet too small")
	ErrNoMorePackets                = errors.New("tsdemuxer: no more packets")
	ErrPacketMustStartWithASyncByte = errors.New("tsdemuxer: packet must start with a sync byte")
)

type TSDemuxer struct {
	ctx context.Context
	r   io.Reader
}

func NewTSDemuxer(ctx context.Context, r io.Reader) *TSDemuxer {
	return &TSDemuxer{
		ctx: ctx,
		r:   r,
	}
}

func (t *TSDemuxer) Demux() error {
	// TODO:
	return nil
}

func (t *TSDemuxer) NextPacket() {
	// TODO:
}
