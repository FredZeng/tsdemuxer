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
}

type TSPacketAdaptationField struct {
	// TODO:
}

func parseTSPacket(data []byte) (p *TSPacket, err error) {
	if len(data) < 4 {
		err = ErrPacketTooSmall
		return
	}

	// check sync_byte
	if data[0] != SyncByte {
		err = ErrPacketMustStartWithASyncByte
		return
	}

	p = &TSPacket{
		Header: parseTSPacketHeader(data[1:4]),
	}

	// TODO:

	return
}

func parseTSPacketHeader(data []byte) TSPacketHeader {
	return TSPacketHeader{
		TransportErrorIndicator:    data[0]&0x80 > 0,
		PayloadUnitStartIndicator:  data[0]&0x40 > 0,
		TransportPriority:          data[0]&0x20 > 0,
		PID:                        uint16(data[0]&0x1f)<<8 | uint16(data[1]),
		TransportScramblingControl: uint8(data[2]) >> 6 & 0x3,
		ContinuityCounter:          uint8(data[2] & 0xf),
		// TODO:
	}
}
