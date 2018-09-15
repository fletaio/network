package network

import (
	"fleta/message"
	"fleta/util"
	"io"
)

type AnswerFormulator struct {
	message.Message
	Addr     string
	NodeType string
}

func NewAnswerFormulator(addr, nodeType string) *AnswerFormulator {
	return &AnswerFormulator{
		Addr:     addr,
		NodeType: nodeType,
	}
}

func (f *AnswerFormulator) WriteTo(w io.Writer) (int64, error) {
	var wrote int64

	num, err := util.WriteUint16(w, uint16(len(f.Addr)))
	if err != nil {
		return wrote, err
	}
	wrote += num

	n, err := w.Write([]byte(f.Addr))
	if err != nil {
		return wrote, err
	}
	wrote += int64(n)

	num, err = util.WriteUint16(w, uint16(len(f.NodeType)))
	if err != nil {
		return wrote, err
	}
	wrote += num

	n, err = w.Write([]byte(f.NodeType))
	if err != nil {
		return wrote, err
	}
	wrote += int64(n)

	return wrote, nil
}

func (f *AnswerFormulator) ReadFrom(r io.Reader) (int64, error) {
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

	Len, n64, err = util.ReadUint16(r)
	if err != nil {
		return read, err
	}
	read += n64
	bs = make([]byte, Len)
	n, err = r.Read(bs)
	if err != nil {
		return read, err
	}
	read += int64(n)
	f.NodeType = string(bs)

	return read, nil
}
