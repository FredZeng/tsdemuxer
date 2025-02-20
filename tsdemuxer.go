package tsdemuxer

import (
	"context"
	"errors"
	"io"
)

const SyncByte = 0x47

var (
	ErrLackOfPacketHeader           = errors.New("tsdemuxer: lack of packet header")
	ErrLackOfAdaptationField        = errors.New("tsdemuxer: lack of adaptation field")
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
