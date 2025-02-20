package tsdemuxer

import "fmt"

type BytesIterator struct {
	bs     []byte
	offset int
}

func NewBytesIterator(bs []byte) *BytesIterator {
	return &BytesIterator{bs: bs}
}

func (i *BytesIterator) NextByte() (b byte, err error) {
	if len(i.bs) < i.offset+1 {
		err = fmt.Errorf("tsdemuxer: slice length is %d, offset %d is invalid", len(i.bs), i.offset)
		return
	}
	b = i.bs[i.offset]
	i.offset++
	return
}

func (i *BytesIterator) NextBytesNoCopy(n int) (bs []byte, err error) {
	if len(i.bs) < i.offset+n {
		err = fmt.Errorf("tsdemuxer: slice length is %d, offset %d is invalid", len(i.bs), i.offset+n)
		return
	}
	bs = i.bs[i.offset : i.offset+n]
	i.offset += n
	return
}

func (i *BytesIterator) Seek(n int) {
	i.offset = n
}

func (i *BytesIterator) Skip(n int) {
	i.offset += n
}

func (i *BytesIterator) Len() int {
	return len(i.bs)
}
