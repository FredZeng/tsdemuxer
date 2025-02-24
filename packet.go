package tsdemuxer

import (
	"fmt"
)

type Packet struct {
	Header          PacketHeader
	AdaptationField *PacketAdaptationField
	Payload         []byte
}

type PacketHeader struct {
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

type PacketAdaptationField struct {
	Length                            int
	DiscontinuityIndicator            bool
	RandomAccessIndicator             bool
	ElementaryStreamPriorityIndicator bool
	PCR                               *ClockReference
	OPCR                              *ClockReference
	SpliceCountdown                   int
	TransportPrivateDataLength        int
	TransportPrivateData              []byte
	AdaptationExtensionField          *PacketAdaptationExtensionField
	HasPCR                            bool
	HasOPCR                           bool
	HasSplicingPoint                  bool
	HasTransportPrivateData           bool
	HasAdaptationExtensionField       bool
}

type PacketAdaptationExtensionField struct {
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

func parsePacket(i *BytesIterator) (p *Packet, err error) {
	var b byte

	// try to get sync_byte
	if b, err = i.NextByte(); err != nil {
		err = fmt.Errorf("tsdemuxer: getting next byte failed: %w", err)
		return
	}

	// check sync_byte (8 bits)
	if b != SyncByte {
		err = ErrPacketMustStartWithASyncByte
		return
	}

	p = &Packet{}

	if p.Header, err = parsePacketHeader(i); err != nil {
		err = fmt.Errorf("tsdemuxer: parsing packet header failed: %w", err)
		return
	}

	if p.Header.HasAdaptationField {
		if p.AdaptationField, err = parsePacketAdaptationField(i); err != nil {
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

func payloadOffset(offsetStart int, h PacketHeader, a *PacketAdaptationField) (offset int) {
	offset = offsetStart + 3
	if h.HasAdaptationField {
		offset += 1 + a.Length
	}
	return
}

func parsePacketHeader(i *BytesIterator) (h PacketHeader, err error) {
	var bs []byte

	if bs, err = i.NextBytesNoCopy(3); err != nil {
		return
	}

	h = PacketHeader{
		TransportErrorIndicator:   bs[0]&0x80 > 0,
		PayloadUnitStartIndicator: bs[0]&0x40 > 0,
		TransportPriority:         bs[0]&0x20 > 0,
		PID:                       uint16(bs[0]&0x1f)<<8 | uint16(bs[1]),
		// 00 - Not scrambled
		// 01 - User-defined
		// 02 - User-defined
		// 03 - User-defined
		TransportScramblingControl: uint8(bs[2]) >> 6 & 0x3,
		// 00 - Reserved for future use by ISO/IEC
		// 01 - No adaptation_field, payload only
		// 10 - Adaptation_field only, no payload
		// 11 - Adaptation_field followed by payload
		AdaptationFieldControl: uint8(bs[2]) >> 4 & 0x3,
		ContinuityCounter:      uint8(bs[2] & 0xf),
		HasAdaptationField:     bs[2]&0x20 > 0,
		HasPayload:             bs[2]&0x10 > 0,
	}

	return
}

func parsePacketAdaptationField(i *BytesIterator) (af *PacketAdaptationField, err error) {
	af = &PacketAdaptationField{}

	var b byte

	// try to get adaptation_field_length
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
			err = fmt.Errorf("tsdemuxer: parsing adaptation field extension failed: %w", err)
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
	cr = newClockReference(int64(pcr>>15), int64(pcr&0x1ff))
	return
}

func parseAdaptationExtensionField(i *BytesIterator) (afe *PacketAdaptationExtensionField, err error) {
	var b byte
	var bs []byte

	afe = &PacketAdaptationExtensionField{}

	// try to get adaptation_field_extension_length
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

	cr = newClockReference(int64(uint64(bs[0])>>1&0x7<<30|uint64(bs[1])<<22|uint64(bs[2])>>1&0x7f<<15|uint64(bs[3])<<7|uint64(bs[4])>>1&0x7f), 0)
	return
}

func (p *Packet) PrettyPrint() {
	fmt.Println("--- Transport Packet ---")
	fmt.Printf("transport_error_indicator: %t\n", p.Header.TransportErrorIndicator)
	fmt.Printf("payload_unit_start_indicator: %t\n", p.Header.PayloadUnitStartIndicator)
	fmt.Printf("transport_priority: %t\n", p.Header.TransportPriority)
	fmt.Printf("PID: %d\n", p.Header.PID)
	fmt.Printf("transport_scrambling_control: %d\n", p.Header.TransportScramblingControl)
	fmt.Printf("adaptation_field_control: %d\n", p.Header.AdaptationFieldControl)
	fmt.Printf("continuity_counter: %d\n", p.Header.ContinuityCounter)

	if p.Header.HasAdaptationField {
		fmt.Println("")
		p.AdaptationField.PrettyPrint()
	}
	fmt.Printf("--- Transport Packet ---\n\n")
}

func (p *PacketAdaptationField) PrettyPrint() {
	fmt.Println("=== Adaptation Field ===")
	fmt.Printf("adaptation_field_length: %d\n", p.Length)
	if p.Length > 0 {
		fmt.Printf("discontinuity_indicator: %t\n", p.DiscontinuityIndicator)
		fmt.Printf("random_access_indicator: %t\n", p.RandomAccessIndicator)
		fmt.Printf("elementary_stream_priority_indicator: %t\n", p.ElementaryStreamPriorityIndicator)
		fmt.Printf("PCR_flag: %t\n", p.HasPCR)
		fmt.Printf("OPCR_flag: %t\n", p.HasOPCR)
		fmt.Printf("splicing_point_flag: %t\n", p.HasSplicingPoint)
		fmt.Printf("transport_private_data_flag: %t\n", p.HasTransportPrivateData)
		fmt.Printf("adaptation_field_extension_flag: %t\n", p.HasAdaptationExtensionField)

		if p.HasPCR {
			fmt.Printf("program_clock_reference_base: %d\n", p.PCR.Base)
			fmt.Printf("program_clock_reference_extension: %d\n", p.PCR.Extension)
		}

		if p.HasOPCR {
			fmt.Printf("original_program_clock_reference_base: %d\n", p.OPCR.Base)
			fmt.Printf("original_program_clock_reference_extension: %d\n", p.OPCR.Extension)
		}

		if p.HasSplicingPoint {
			fmt.Printf("splice_countdown: %d\n", p.SpliceCountdown)
		}

		if p.HasTransportPrivateData {
			fmt.Printf("transport_private_data_length: %d\n", p.TransportPrivateDataLength)
			//fmt.Printf("transport_private_data: %v\n", p.TransportPrivateData)
		}

		if p.HasAdaptationExtensionField {
			fmt.Printf("adaptation_extension_field_length: %d\n", p.AdaptationExtensionField.Length)
			fmt.Printf("legal_time_window_flag: %t\n", p.AdaptationExtensionField.HasLegalTimeWindow)
			fmt.Printf("piecewise_rate_flag: %t\n", p.AdaptationExtensionField.HasPiecewiseRate)
			fmt.Printf("seamless_splice_flag: %t\n", p.AdaptationExtensionField.HasSeamlessSplice)

			if p.AdaptationExtensionField.HasLegalTimeWindow {
				fmt.Printf("legal_time_window_valid_flag: %t\n", p.AdaptationExtensionField.LegalTimeWindowIsValid)
				fmt.Printf("legal_time_window_offset: %d\n", p.AdaptationExtensionField.LegalTimeWindowOffset)
			}

			if p.AdaptationExtensionField.HasPiecewiseRate {
				fmt.Printf("piecewise_rate: %d\n", p.AdaptationExtensionField.PiecewiseRate)
			}

			if p.AdaptationExtensionField.HasSeamlessSplice {
				fmt.Printf("splice_type: %d\n", p.AdaptationExtensionField.SpliceType)
				fmt.Printf("DTS_next_Access_Unit: %d\n", p.AdaptationExtensionField.DTSNextAccessUnit.Base)
			}
		}
	}
	fmt.Println("=== Adaptation Field ===")
}
