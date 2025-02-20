package tsdemuxer

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
	AdaptationExtensionField          *PacketAdaptationExtensionField
	HasPCR                            bool
	HasOPCR                           bool
	HasSplicingPoint                  bool
	HasTransportPrivateData           bool
	HasAdaptationExtensionField       bool
}

type ClockReference struct {
	Base      int64
	Extension int64
}

type PacketAdaptationExtensionField struct {
	Length int
	// TODO:
}

func parseTSPacket(data *BytesIterator) (p *TSPacket, err error) {
	if data.Len() < 4 {
		err = ErrLackOfPacketHeader
		return
	}

	// check sync_byte
	if b, _ := data.NextByte(); b != SyncByte {
		err = ErrPacketMustStartWithASyncByte
		return
	}

	p = &TSPacket{
		Header: parseTSPacketHeader(data),
	}

	if p.Header.HasAdaptationField {
		if p.AdaptationField, err = parseTSPacketAdaptationField(data); err != nil {
			err = ErrLackOfAdaptationField
			return
		}
	}

	// TODO:

	return
}

func parseTSPacketHeader(data *BytesIterator) TSPacketHeader {
	b, _ := data.NextBytesNoCopy(3)

	return TSPacketHeader{
		TransportErrorIndicator:   b[0]&0x80 > 0,
		PayloadUnitStartIndicator: b[0]&0x40 > 0,
		TransportPriority:         b[0]&0x20 > 0,
		PID:                       uint16(b[0]&0x1f)<<8 | uint16(b[1]),
		// 00 - Not scrambled
		// 01 - User-defined
		// 02 - User-defined
		// 03 - User-defined
		TransportScramblingControl: uint8(b[2]) >> 6 & 0x3,
		// 00 - Reserved for future use by ISO/IEC
		// 01 - No adaptation_field, payload only
		// 10 - Adaptation_field only, no payload
		// 11 - Adaptation_field followed by payload
		AdaptationFieldControl: uint8(b[2]) >> 4 & 0x3,
		ContinuityCounter:      uint8(b[2] & 0xf),
		HasAdaptationField:     b[2]&0x20 > 0,
		HasPayload:             b[2]&0x10 > 0,
	}
}

func parseTSPacketAdaptationField(data *BytesIterator) (a *TSPacketAdaptationField, err error) {
	a = &TSPacketAdaptationField{}

	var b byte

	if b, err = data.NextByte(); err != nil {
		return
	}

	a.Length = int(b)

	if a.Length == 0 {
		return
	}

	// FIXME:
	//if len(data) < a.Length || len(data) < 2 {
	//	err = ErrLackOfAdaptationField
	//	return
	//}
	//
	//a.DiscontinuityIndicator = data[1]&0x80 > 0
	//a.RandomAccessIndicator = data[1]&0x40 > 0
	//a.ElementaryStreamPriorityIndicator = data[1]&0x20 > 0
	//a.HasPCR = data[1]&0x10 > 0
	//a.HasOPCR = data[1]&0x08 > 0
	//a.HasSplicingPoint = data[1]&0x04 > 0
	//a.HasTransportPrivateData = data[1]&0x02 > 0
	//a.HasAdaptationExtensionField = data[1]&0x01 > 0
	//
	//if a.HasPCR {
	//	a.PCR = parsePCR(data[2:8])
	//}
	//
	//if a.HasOPCR {
	//	a.OPCR = parsePCR(data[8:16])
	//}

	return
}

func parsePCR(data []byte) *ClockReference {
	// TODO:
	return nil
}
