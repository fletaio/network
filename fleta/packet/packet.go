package packet

import (
	"bytes"
	"compress/gzip"
	"io"

	"fleta/encoding"
)

type Packet struct {
	Header  Header
	Payload Payload
}

func NewSendPacket(payload Payload, compression CompressionType) *Packet {
	return &Packet{
		Header: Header{
			Magicword:   MAGICWORD,
			Compression: compression,
		},
		Payload: payload,
	}
}

func NewRecvPacket() *Packet {
	return new(Packet)
}

func (packet *Packet) WriteTo(w io.Writer) (int64, error) {
	var wrote int64

	buf := &bytes.Buffer{}

	if packet.Header.Compression == COMPRESSED {
		// TODO
		gw := gzip.NewWriter(buf)
		if _, err := packet.Payload.WriteTo(gw); err != nil {
			return wrote, err
		}
		if err := gw.Flush(); err != nil {
			return wrote, err
		}
		if err := gw.Close(); err != nil {
			return wrote, err
		}
	} else {
		packet.Payload.WriteTo(buf)
	}
	packet.Payload = buf
	packet.Header.Size = buf.Len()

	if n, err := packet.Header.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}

	integrity := new(encoding.CRC32)
	e := encoding.NewWriter(w, integrity)

	if n, err := packet.Payload.WriteTo(e); err != nil {
		return wrote, err
	} else {
		wrote += n
	}

	if n, err := integrity.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}

	return wrote, nil
}

func (packet *Packet) ReadFrom(r io.Reader) (int64, error) {
	var read int64

	packet.Header = Header{}
	if n, err := packet.Header.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}

	integrity := new(encoding.CRC32)
	e, err := encoding.NewReader(r, integrity)
	if err != nil {
		return read, err
	}

	bs := make([]byte, packet.Header.Size)
	if n, err := e.Read(bs); n != packet.Header.Size {
		return read, ErrMismatchPacketSize
	} else if err != nil {
		return read, err
	} else {
		read += int64(n)
	}

	if n, err := integrity.ReadFrom(r); err != nil {
		read += n
		return read, err
	} else {
		read += n
	}

	if packet.Header.Compression == COMPRESSED {
		var buf bytes.Buffer
		gr, err := gzip.NewReader(bytes.NewBuffer(bs))
		defer gr.Close()
		_, err = buf.ReadFrom(gr)
		if err != nil {
			return read, err
		}
		packet.Payload = &buf
	} else {
		packet.Payload = bytes.NewBuffer(bs)
	}

	return read, nil
}

func (packet *Packet) GetPayload() Payload {
	return packet.Payload
}
