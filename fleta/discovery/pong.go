package discovery

import (
	"fleta/message"
	"fleta/util"
	"io"
)

type Pong struct {
	message.Message
	time  uint32
	test  uint64
	test2 uint8
}

func (pong *Pong) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := util.WriteUint32(w, pong.time); err != nil {
		return wrote, nil
	} else {
		wrote += n
	}

	if n, err := util.WriteUint64(w, pong.test); err != nil {
		return wrote, nil
	} else {
		wrote += n
	}

	if n, err := util.WriteUint8(w, pong.test2); err != nil {
		return wrote, nil
	} else {
		wrote += n
	}

	return wrote, nil
}

func (pong *Pong) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	if v, n, err := util.ReadUint32(r); err != nil {
		return read, nil
	} else {
		read += n
		pong.time = v
	}

	if v, n, err := util.ReadUint64(r); err != nil {
		return read, nil
	} else {
		read += n
		pong.test = v
	}

	if v, n, err := util.ReadUint8(r); err != nil {
		return read, nil
	} else {
		read += n
		pong.test2 = v
	}

	return read, nil
}
