package network

import (
	"io"
)

// Reader for tolerance in network conn
type Reader struct {
	reader io.Reader
}

func NewReader(r io.Reader) *Reader {
	return &Reader{
		reader: r,
	}
}

func (pr *Reader) Read(bs []byte) (int, error) {
	return io.ReadFull(pr.reader, bs)
}
