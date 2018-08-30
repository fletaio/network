package peer

import (
	"fleta/message"
	"net"
)

type NodeID string

type IPeer interface {
	Bond(conn net.Conn) chan error
	GetConn() *PeerConn
	CheckConn() bool
}

type Peer struct {
	ID              NodeID
	conn            *PeerConn
	MessageResolver message.Resolver
	Errc            chan error
}

func NewPeer(id NodeID, resolver message.Resolver) *Peer {
	return &Peer{
		ID:              id,
		MessageResolver: resolver,
		Errc:            make(chan error),
	}
}

func (p *Peer) CheckConn() bool {
	if p == nil {
		return false
	}
	if p.conn == nil {
		return false
	}
	return p.conn.CheckConn()
}

func (p *Peer) Bond(conn net.Conn) chan error {
	p.conn = NewPeerConn(conn, p.MessageResolver, p.Errc)
	p.conn.Run()
	return p.Errc
}

func (p *Peer) GetConn() *PeerConn {
	return p.conn
}
