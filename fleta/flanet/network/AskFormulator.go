package network

import (
	"fleta/message"
	"fleta/util"
	"io"
)

type AskFormulator struct {
	message.Message
	Addr string
}

func NewAskFormulator(addr string) *AskFormulator {
	return &AskFormulator{
		Addr: addr,
	}
}

func (f *AskFormulator) WriteTo(w io.Writer) (int64, error) {
	var wrote int64

	bAddr := []byte(f.Addr)
	num, err := util.WriteUint16(w, uint16(len(bAddr)))
	if err != nil {
		return wrote, err
	}
	wrote += num

	n, err := w.Write(bAddr)
	if err != nil {
		return wrote, err
	}
	wrote += int64(n)

	return wrote, nil
}

func (f *AskFormulator) ReadFrom(r io.Reader) (int64, error) {
	var read int64

	Len, n64, err := util.ReadUint16(r)
	if err != nil {
		return read, err
	}
	read += n64
	bs := make([]byte, Len)
	n, err := r.Read(bs)
	if err != nil {
		return read, err
	}
	read += int64(n)
	f.Addr = string(bs)

	return read, nil
}
