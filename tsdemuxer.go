package tsdemuxer

import (
	"context"
	"io"
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
