package network

import (
	"net"
	"strconv"
	"sync"
	"time"

	stringutil "github.com/fletaio/network/string_util"
)

type mockAddr struct {
	network string
	address string
	port    int
}

func (c *mockAddr) Network() string {
	return c.network
}
func (c *mockAddr) String() string {
	return c.address
}

func mockDelay(address string, target string) time.Duration {
	address = stringutil.Sha256HexString2(address)
	target = stringutil.Sha256HexString2(target)

	a, _ := strconv.ParseInt(string(target[0]), 16, 64)
	b, _ := strconv.ParseInt(string(address[0]), 16, 64)
	length := int(((a+b)*DelayUnit)/32) + 1

	delay := time.Duration(0)
	for i := 0; i < length; i++ {
		a, _ := strconv.ParseInt(string(target[i]), 16, 64)
		b, _ := strconv.ParseInt(string(address[i]), 16, 64)
		delay += time.Duration(a+b) / 2
	}

	return time.Millisecond * delay
}

// Conn is a generic stream-oriented network connection.
// type Conn net.Conn

type mockConn struct {
	Conn          net.Conn
	LocalAddrVal  mockAddr
	RemoteAddrVal mockAddr
	readDeadline  time.Duration
	writeDeadline time.Duration
}

func (c *mockConn) Read(b []byte) (n int, err error) {
	return c.Conn.Read(b)
}

func (c *mockConn) Write(b []byte) (n int, err error) {
	if Delay {
		delay := mockDelay(c.LocalAddr().String(), c.RemoteAddr().String())
		time.Sleep(delay)
	}

	n, err = c.Conn.Write(b)

	return
}

func (c *mockConn) Close() error {
	c.Conn.Close()
	return nil
}
func (c *mockConn) LocalAddr() net.Addr {
	return &c.LocalAddrVal
}
func (c *mockConn) RemoteAddr() net.Addr {
	return &c.RemoteAddrVal
}
func (c *mockConn) SetDeadline(t time.Time) error {
	c.readDeadline = time.Since(t)
	c.writeDeadline = time.Since(t)
	return nil
}
func (c *mockConn) SetReadDeadline(t time.Time) error {
	c.readDeadline = time.Since(t)
	return nil
}
func (c *mockConn) SetWriteDeadline(t time.Time) error {
	c.writeDeadline = time.Since(t)
	return nil
}

var connLock sync.Mutex

func getConnPair() (net.Conn, net.Conn) {
	// return net.Pipe()
	connLock.Lock()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	var s net.Conn
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		s, err = l.Accept()
		if err != nil {
			panic(err)
		}
		wg.Done()
	}()
	c, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		panic(err)
	}
	wg.Wait()
	l.Close()
	connLock.Unlock()

	return s, c
}
