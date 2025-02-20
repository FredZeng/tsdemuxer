package tsdemuxer

import (
	"fmt"
)

type TSPacket struct {
	Header          TSPacketHeader
	AdaptationField *TSPacketAdaptationField
	Payload         []byte
}

type TSPacketHeader struct {
	TransportErrorIndicator    bool
	PayloadUnitStartIndicator  bool
	TransportPriority          bool
	PID                        uint16
	TransportScramblingControl uint8
	AdaptationFieldControl     uint8
	ContinuityCounter          uint8
	HasAdaptationField         bool
	HasPayload                 bool
}

type TSPacketAdaptationField struct {
	Length                            int
	DiscontinuityIndicator            bool
	RandomAccessIndicator             bool
	ElementaryStreamPriorityIndicator bool
	PCR                               *ClockReference
	OPCR                              *ClockReference
	SpliceCountdown                   int
	TransportPrivateDataLength        int
	TransportPrivateData              []byte
	AdaptationExtensionField          *TSPacketAdaptationExtensionField
	HasPCR                            bool
	HasOPCR                           bool
	HasSplicingPoint                  bool
	HasTransportPrivateData           bool
	HasAdaptationExtensionField       bool
}

type TSPacketAdaptationExtensionField struct {
	Length                 int
	HasLegalTimeWindow     bool
	HasPiecewiseRate       bool
	HasSeamlessSplice      bool
	LegalTimeWindowIsValid bool
	LegalTimeWindowOffset  uint16
	PiecewiseRate          uint32
	SpliceType             uint8
	DTSNextAccessUnit      *ClockReference
}

func parseTSPacket(i *BytesIterator) (p *TSPacket, err error) {
	var b byte

	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("tsdemuxer: getting next byte failed: %w", err)
		return
	}

	// check sync_byte (8 bits)
	if b != SyncByte {
		err = ErrPacketMustStartWithASyncByte
		return
	}

	p = &TSPacket{}

	if p.Header, err = parseTSPacketHeader(i); err != nil {
		err = fmt.Errorf("tsdemuxer: parsing packet header failed: %w", err)
		return
	}

	if p.Header.HasAdaptationField {
		if p.AdaptationField, err = parseTSPacketAdaptationField(i); err != nil {
			err = fmt.Errorf("tsdemuxer: parsing packet adaptation field failed: %w", err)
			return
		}
	}

	if p.Header.HasPayload {
		i.Seek(payloadOffset(1, p.Header, p.AdaptationField))
		p.Payload = i.Dump()
	}

	return
}

func payloadOffset(offsetStart int, h TSPacketHeader, a *TSPacketAdaptationField) (offset int) {
	offset = offsetStart + 3
	if h.HasAdaptationField {
		offset += 1 + a.Length
	}
	return
}

func parseTSPacketHeader(i *BytesIterator) (h TSPacketHeader, err error) {
	var bytes []byte

	if bytes, err = i.NextBytesNoCopy(3); err != nil {
		return
	}

	h = TSPacketHeader{
		TransportErrorIndicator:   bytes[0]&0x80 > 0,
		PayloadUnitStartIndicator: bytes[0]&0x40 > 0,
		TransportPriority:         bytes[0]&0x20 > 0,
		PID:                       uint16(bytes[0]&0x1f)<<8 | uint16(bytes[1]),
		// 00 - Not scrambled
		// 01 - User-defined
		// 02 - User-defined
		// 03 - User-defined
		TransportScramblingControl: uint8(bytes[2]) >> 6 & 0x3,
		// 00 - Reserved for future use by ISO/IEC
		// 01 - No adaptation_field, payload only
		// 10 - Adaptation_field only, no payload
		// 11 - Adaptation_field followed by payload
		AdaptationFieldControl: uint8(bytes[2]) >> 4 & 0x3,
		ContinuityCounter:      uint8(bytes[2] & 0xf),
		HasAdaptationField:     bytes[2]&0x20 > 0,
		HasPayload:             bytes[2]&0x10 > 0,
	}

	return
}

func parseTSPacketAdaptationField(i *BytesIterator) (af *TSPacketAdaptationField, err error) {
	af = &TSPacketAdaptationField{}

	var b byte

	if b, err = i.NextByte(); err != nil {
		return
	}

	af.Length = int(b)

	if af.Length == 0 {
		return
	}

	if b, err = i.NextByte(); err != nil {
		return
	}

	af.DiscontinuityIndicator = b&0x80 > 0
	af.RandomAccessIndicator = b&0x40 > 0
	af.ElementaryStreamPriorityIndicator = b&0x20 > 0
	af.HasPCR = b&0x10 > 0
	af.HasOPCR = b&0x08 > 0
	af.HasSplicingPoint = b&0x04 > 0
	af.HasTransportPrivateData = b&0x02 > 0
	af.HasAdaptationExtensionField = b&0x01 > 0

	if af.HasPCR {
		if af.PCR, err = parsePCR(i); err != nil {
			err = fmt.Errorf("tsdemuxer: parsing PCR failed: %w", err)
			return
		}
	}

	if af.HasOPCR {
		if af.OPCR, err = parsePCR(i); err != nil {
			err = fmt.Errorf("tsdemuxer: parsing OPCR failed: %w", err)
			return
		}
	}

	if af.HasSplicingPoint {
		if b, err = i.NextByte(); err != nil {
			err = fmt.Errorf("tsdemuxer: getting next byte failed: %w", err)
			return
		}

		af.SpliceCountdown = int(b)
	}

	if af.HasTransportPrivateData {
		if b, err = i.NextByte(); err != nil {
			err = fmt.Errorf("tsdemuxer: getting next byte failed: %w", err)
			return
		}
		af.TransportPrivateDataLength = int(b)

		if af.TransportPrivateDataLength > 0 {
			if af.TransportPrivateData, err = i.NextBytes(af.TransportPrivateDataLength); err != nil {
				err = fmt.Errorf("tsdemuxer: getting next bytes failed: %w", err)
				return
			}
		}
	}

	if af.HasAdaptationExtensionField {
		if af.AdaptationExtensionField, err = parseAdaptationExtensionField(i); err != nil {
			err = fmt.Errorf("tsdemuxer: parsing adaptation field extension: %w", err)
			return
		}
	}

	return
}

func parsePCR(i *BytesIterator) (cr *ClockReference, err error) {
	var bs []byte

	if bs, err = i.NextBytesNoCopy(6); err != nil {
		return
	}

	pcr := uint64(bs[0])<<40 | uint64(bs[1])<<32 | uint64(bs[2])<<24 | uint64(bs[3])<<16 | uint64(bs[4])<<8 | uint64(bs[5])
	cr = NewClockReference(int64(pcr>>15), int64(pcr&0x1ff))
	return
}

func parseAdaptationExtensionField(i *BytesIterator) (afe *TSPacketAdaptationExtensionField, err error) {
	var b byte
	var bs []byte

	afe = &TSPacketAdaptationExtensionField{}

	if b, err = i.NextByte(); err != nil {
		return
	}

	afe.Length = int(b)

	if afe.Length > 0 {
		if b, err = i.NextByte(); err != nil {
			return
		}

		afe.HasLegalTimeWindow = b&0x80 > 0
		afe.HasPiecewiseRate = b&0x40 > 0
		afe.HasSeamlessSplice = b&0x20 > 0

		if afe.HasLegalTimeWindow {
			if bs, err = i.NextBytesNoCopy(2); err != nil {
				return
			}

			afe.LegalTimeWindowIsValid = bs[0]&0x80 > 0
			afe.LegalTimeWindowOffset = uint16(bs[0]&0x7f)<<8 | uint16(bs[1])
		}

		if afe.HasPiecewiseRate {
			if bs, err = i.NextBytesNoCopy(3); err != nil {
				return
			}

			afe.PiecewiseRate = uint32(bs[0]&0x3f)<<16 | uint32(bs[1])<<8 | uint32(bs[2])
		}

		if afe.HasSeamlessSplice {
			if b, err = i.NextByte(); err != nil {
				return
			}

			afe.SpliceType = uint8(b&0xf0) >> 4

			i.Skip(-1)

			if afe.DTSNextAccessUnit, err = parsePTSOrDTS(i); err != nil {
				err = fmt.Errorf("tsdemuxer: parsing PTS or DTS failed: %w", err)
				return
			}
		}
	}

	return
}

func parsePTSOrDTS(i *BytesIterator) (cr *ClockReference, err error) {
	var bs []byte

	if bs, err = i.NextBytesNoCopy(5); err != nil {
		return
	}

	cr = NewClockReference(int64(uint64(bs[0])>>1&0x7<<30|uint64(bs[1])<<22|uint64(bs[2])>>1&0x7f<<15|uint64(bs[3])<<7|uint64(bs[4])>>1&0x7f), 0)
	return
}
