package tsdemuxer

import (
	"context"
	"errors"
	"fmt"
	"io"
)

const SyncByte = 0x47
const TSPacketSize = 188

const (
	PAT_ID  uint16 = 0x0000
	NULL_ID uint16 = 0x1FFF
)

var (
	ErrNoMorePackets                = errors.New("tsdemuxer: no more packets")
	ErrPacketMustStartWithASyncByte = errors.New("tsdemuxer: packet must start with a sync byte")
)

type Demuxer struct {
	ctx          context.Context
	r            io.Reader
	packetBuffer *packetBuffer
}

type Frame struct {
	streamType uint8
	payload    []byte
	pts        int64
	dts        int64
}

func NewDemuxer(ctx context.Context, r io.Reader) *Demuxer {
	return &Demuxer{
		ctx: ctx,
		r:   r,
	}
}

func (d *Demuxer) NextPacket() (p *Packet, err error) {
	if err = d.ctx.Err(); err != nil {
		return
	}

	if d.packetBuffer == nil {
		d.packetBuffer = newPacketBuffer(d.r, TSPacketSize)
	}

	if p, err = d.packetBuffer.next(); err != nil {
		if !errors.Is(err, ErrNoMorePackets) {
			err = fmt.Errorf("tsdemuxer: getting next packet from buffer failed: %w", err)
		}
		return
	}

	p.PrettyPrint()

	return
}

func (d *Demuxer) NextFrame() (f *Frame, err error) {
	for {
		var p *Packet
		if p, err = d.NextPacket(); err != nil {
			return
		}

		if p.Header.PID == PAT_ID {
			// TODO:
		} else if p.Header.PID == NULL_ID {
			continue
		} else {
			// TODO:
		}
	}

	return
}
