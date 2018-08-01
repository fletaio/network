package packet

// type Packet struct {
// 	Header Header
// 	r io.Reader
// }

// func (p *Packet) ReadFrom(r io.Reader) {
// 	//TODO : read header
// 	p.r = r
// }

// func (p *Packet) Read(bs []byte) (int, error) {
// 	if n, err := p.r.Read(bs); err != nil {
// 		return n, err
// 	} else {
// 		//TODO : update CRC
// 	}
// 	return 0, nil
// }

// func Test() error {
// 	payload, err := packet.GetPayload(p.conn)
// 	if err != nil {
// 		//TODO
// 	}
// 	if len(payload) < 2 {
// 		return ErrInvalidMessagePayload
// 	}

// 	bProcess := false
// 	t := MessageType(BytesToUint16(payload[:2]))
// 	if msg, h, err := pc.MessageResolver.Resolve(t, bytes.NewReader(payload[2:])); err != nil {
// 		if err != ErrUnknownMessage {
// 			return err
// 		}
// 	} else {
// 		if err := p.IsValidCRC(); err != nil {
// 			return err
// 		}
// 		if err := h.Process(msg); err != nil {
// 			return err
// 		}
// 		bProcess = true
// 		break
// 	}
// }

// func ReadPacket(r io.Reader) (*Packet, error) {

// }
