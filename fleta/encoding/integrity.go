package encoding

import (
	"hash/crc32"
	"io"

	"fleta/util"
)

var IEEETable = crc32.MakeTable(crc32.IEEE)

type (
	// TODO: Generalize
	Integrity interface {
		io.ReaderFrom
		io.WriterTo
		Result() uint32
		Update(bs []byte)
	}

	CRC32 struct {
		checksum uint32
		read     uint64
	}
)

func (c *CRC32) Result() uint32 {
	return c.checksum
}

func (c *CRC32) Update(bs []byte) {
	if c.read == 0 {
		c.checksum = crc32.Checksum(bs, IEEETable)
	} else {
		c.checksum = crc32.Update(c.checksum, IEEETable, bs)
	}
	c.read += uint64(len(bs))
}

func (c *CRC32) WriteTo(w io.Writer) (int64, error) {
	n, err := util.WriteUint32(w, c.checksum)
	if err != nil {
		return n, err
	}
	return n, nil
}

func (c *CRC32) ReadFrom(r io.Reader) (int64, error) {
	checksum, n, err := util.ReadUint32(r)
	if err != nil {
		return n, err
	}
	if checksum != c.checksum {
		return n, ErrInvalidIntegrity
	}

	return n, nil
}
