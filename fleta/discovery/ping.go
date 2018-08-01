package discovery

import (
	"fleta/message"
	"fleta/util"
	"io"
)

type Ping struct {
	message.Message
	time  uint32
	test  uint64
	test2 uint8
}

func NewPing(time uint32, test uint64, test2 uint8) *Ping {
	return &Ping{
		time:  time,
		test:  test,
		test2: test2,
	}
}

func (ping *Ping) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := util.WriteUint32(w, ping.time); err != nil {
		return wrote, nil
	} else {
		wrote += n
	}

	if n, err := util.WriteUint64(w, ping.test); err != nil {
		return wrote, nil
	} else {
		wrote += n
	}

	if n, err := util.WriteUint8(w, ping.test2); err != nil {
		return wrote, nil
	} else {
		wrote += n
	}

	return wrote, nil
}

func (ping *Ping) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	if v, n, err := util.ReadUint32(r); err != nil {
		return read, nil
	} else {
		read += n
		ping.time = v
	}

	if v, n, err := util.ReadUint64(r); err != nil {
		return read, nil
	} else {
		read += n
		ping.test = v
	}

	if v, n, err := util.ReadUint8(r); err != nil {
		return read, nil
	} else {
		read += n
		ping.test2 = v
	}

	return read, nil
}
