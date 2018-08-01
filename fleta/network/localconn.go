package network

import (
	"bytes"
	"net"
	"sync"
)

type LocalConn struct {
	net.Conn
	Buf *bytes.Buffer
	m   sync.Mutex
}

func NewLocalConn() *LocalConn {
	return &LocalConn{
		Buf: &bytes.Buffer{},
	}
}

func (c *LocalConn) Read(bs []byte) (int, error) {
	c.m.Lock()
	defer c.m.Unlock()
	var n int
	for {
		var err error
		n, err = c.Buf.Read(bs)

		if err != nil {
			return n, err
		} else {
			break
		}
	}
	return n, nil
}

func (c *LocalConn) Write(bs []byte) (int, error) {
	c.m.Lock()
	defer c.m.Unlock()
	return c.Buf.Write(bs)
}

func (c *LocalConn) Close() error {
	c.Buf = nil
	return nil
}
