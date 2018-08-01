package message

import (
	"encoding/binary"
	"io"
	"log"

	"fleta/packet"
	"fleta/util"
)

// MessageType TODO
type MessageType uint64

func DefineType(t string) MessageType {
	if l := len(t); l > 8 {
		log.Panic("Check")
	} else if l < 8 {
		bs := make([]byte, 8)
		copy(bs[:], t)
		return MessageType(binary.BigEndian.Uint64(bs))
	}
	return MessageType(binary.BigEndian.Uint64([]byte(t)))
}

type Message interface {
	io.WriterTo
	io.ReaderFrom
}

type MessagePayload struct {
	packet.Payload
	MsgType MessageType
	Msg     Message
}

func ToPayload(msgType MessageType, msg Message) packet.Payload {
	return &MessagePayload{
		MsgType: msgType,
		Msg:     msg,
	}
}

func (mp *MessagePayload) WriteTo(w io.Writer) (int64, error) {
	var wrote int64

	if n, err := util.WriteUint64(w, uint64(mp.MsgType)); err != nil {
		return wrote, err
	} else {
		wrote += n
	}

	if n, err := mp.Msg.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}

	return wrote, nil
}

func (mp *MessagePayload) Len() int {
	return -1
}

func Handle(payload packet.Payload, resolver Resolver) (int64, error) {
	var read int64
	v, n, err := util.ReadUint64(payload)
	if err != nil {
		return read, err
	} else {
		read += n
	}

	msg, p, n, err := resolver.Resolve(MessageType(v), payload)
	if err != nil {
		return read, err
	} else {
		read += n
	}

	err = p.Process(msg)

	if err != nil {
		return read, err
	}

	return read, nil
}
