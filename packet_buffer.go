package tsdemuxer

import (
	"errors"
	"fmt"
	"io"
)

type packetBuffer struct {
	packetSize       int
	r                io.Reader
	packetReadBuffer []byte
}

func newPacketBuffer(r io.Reader, packetSize int) *packetBuffer {
	return &packetBuffer{
		packetSize: packetSize,
		r:          r,
	}
}

func (pb *packetBuffer) next() (p *Packet, err error) {
	// Read
	if pb.packetReadBuffer == nil || len(pb.packetReadBuffer) != pb.packetSize {
		pb.packetReadBuffer = make([]byte, pb.packetSize)
	}

	for p == nil {
		if _, err = io.ReadFull(pb.r, pb.packetReadBuffer); err != nil {
			if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
				err = ErrNoMorePackets
			} else {
				err = fmt.Errorf("tsdemuxer: reading %d bytes failed: %w", pb.packetSize, err)
			}
			return
		}

		if p, err = parsePacket(NewBytesIterator(pb.packetReadBuffer)); err != nil {
			err = fmt.Errorf("tsdemuxer: building packet failed: %w", err)
			return
		}
	}
	return
}
