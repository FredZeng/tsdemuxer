package tsdemuxer

type ClockReference struct {
	Base      int64
	Extension int64
}

func NewClockReference(base, extension int64) *ClockReference {
	return &ClockReference{
		Base:      base,
		Extension: extension,
	}
}
