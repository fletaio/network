package peer

import (
	"io"
	"net"
	"sync"

	"fleta/message"
	"fleta/mock/simulationlog"
	"fleta/network"
	"fleta/packet"
)

type PeerConn struct {
	conn            net.Conn
	MessageResolver message.Resolver
	errc            chan error
	mSend           sync.Mutex
}

func NewPeerConn(conn net.Conn, resolver message.Resolver, errc chan error) *PeerConn {
	return &PeerConn{
		conn:            conn,
		MessageResolver: resolver,
		errc:            errc,
	}
}

func (p *PeerConn) Send(payload *packet.Payload, compression packet.CompressionType) error {
	p.mSend.Lock()
	simulationlog.Send(p.conn.LocalAddr().String(), p.conn.RemoteAddr().String(), *payload)
	_, err := Send(p.conn, *payload, compression)
	p.mSend.Unlock()
	return err
}

func Send(conn io.Writer, payload packet.Payload, compression packet.CompressionType) (int64, error) {
	packet := packet.NewSendPacket(payload, compression)
	return packet.WriteTo(conn)
}

func (p *PeerConn) CheckConn() bool {
	if p.conn != nil {
		return true
	}
	return false
}

func (p *PeerConn) Close() error {
	return p.conn.Close()
}

func (p *PeerConn) Run() {
	go p.recvLoop()
}

func (p *PeerConn) recvLoop() {
	var err error

	for {
		_, err = Recv(p.conn, p.MessageResolver)
		if err != nil {
			break
		}
	}

	p.conn.Close()
	p.errc <- err
}

func Recv(conn io.Reader, resolver message.Resolver) (int64, error) {

	packet := packet.NewRecvPacket()

	readall := network.NewReader(conn)
	n, err := packet.ReadFrom(readall)
	if err != nil {
		return n, err
	}

	payload := packet.GetPayload()
	plen := int64(payload.Len())
	hlen, err := message.Handle(payload, resolver)
	if err != nil {
		return n, err
	} else if hlen != plen {
		return n, message.ErrInvalidMsgConsumeSize
	}

	return n, nil
}
