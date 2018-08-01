package packet

import (
	"io"

	"fleta/util"
)

const MAGICWORD = 'F'

type CompressionType uint8

const (
	UNCOMPRESSED = CompressionType(0)
	COMPRESSED   = CompressionType(1)
)

// Header TODO
type Header struct {
	Magicword   uint8
	Compression CompressionType
	Size        int
}

// WriteTo TODO
func (header *Header) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := util.WriteUint8(w, header.Magicword); err != nil {
		return wrote, err
	} else {
		wrote += n
	}

	if n, err := util.WriteUint8(w, uint8(header.Compression)); err != nil {
		return wrote, err
	} else {
		wrote += n
	}

	if n, err := util.WriteUint32(w, uint32(header.Size)); err != nil {
		return wrote, err
	} else {
		wrote += n
	}

	return wrote, nil
}

// ReadFrom TODO
func (header *Header) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	if v, n, err := util.ReadUint8(r); err != nil {
		return read, err
	} else if v != MAGICWORD {
		return read, ErrMismatchMagicword
	} else {
		read += n
		header.Magicword = v
	}

	if v, n, err := util.ReadUint8(r); err != nil {
		return read, err
	} else {
		read += n
		header.Compression = CompressionType(v)
	}

	if v, n, err := util.ReadUint32(r); err != nil {
		return read, err
	} else {
		read += n
		header.Size = int(v)
	}

	return read, nil
}
