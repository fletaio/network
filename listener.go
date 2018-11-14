package mocknet

import (
	"net"
	"sync"
)

type mockListener struct {
	sync.Mutex
	addr  net.Addr
	node  NodeInfo
	count int
}

func (l *mockListener) waitAccept() {
	l.node = registAccept(l.addr.String())
}

// Accept waits for and returns the next connection to the listener.
func (l *mockListener) Accept() (net.Conn, error) {
	connParam := <-l.node.ConnParamChan

	mockconn := mockConn{
		readDeadline:  -1,
		writeDeadline: -1,
	}
	mockconn.Conn = connParam.Conn
	mockconn.RemoteAddrVal = mockAddr{
		network: connParam.NetworkType,
		address: connParam.DialHost,
	}
	mockconn.LocalAddrVal = mockAddr{
		network: connParam.NetworkType,
		address: connParam.Address,
	}

	var c net.Conn
	c = &mockconn
	return c, nil
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (l *mockListener) Close() error {
	return nil
}

// Addr returns the listener's network address.
func (l *mockListener) Addr() net.Addr {
	return net.Addr(l.addr)
}
