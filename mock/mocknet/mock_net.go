package mocknet

import (
	"context"
	"errors"
	"net"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"git.fleta.io/fleta/framework/log"
	"git.fleta.io/fleta/mock"
	"git.fleta.io/fleta/mock/mocknetwork"
	"git.fleta.io/fleta/mock/simulationlog"
	"git.fleta.io/fleta/samutil"
)

var myID string
var (
	//ErrDialTimeout is timeout error
	ErrDialTimeout = errors.New("Dial timeout error")
)

type timeoutError struct {
	error
}

func (t *timeoutError) Timeout() bool {
	return true
}
func (t *timeoutError) Temporary() bool {
	return false
}

type imockAddr interface {
	Network() string
	String() string
}
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

type mockConn struct {
	Conn          net.Conn
	LocalAddrVal  mockAddr
	RemoteAddrVal mockAddr
	targetID      string
	readDeadline  time.Duration
	writeDeadline time.Duration
}

func mockDelay(address string, target string) time.Duration {
	address = samutil.Sha256HexString2(address)
	target = samutil.Sha256HexString2(target)

	a, _ := strconv.ParseInt(string(target[0]), 16, 64)
	b, _ := strconv.ParseInt(string(address[0]), 16, 64)
	length := int(((a+b)*simulationdata.DelayUnit)/32) + 1

	delay := time.Duration(0)
	for i := 1; i < length; i++ {
		a, _ := strconv.ParseInt(string(target[i]), 16, 64)
		b, _ := strconv.ParseInt(string(address[i]), 16, 64)
		delay += time.Duration(a+b) / 2
	}

	return time.Millisecond * delay
}

func (c *mockConn) Log(format string, msg ...interface{}) {
	buf := make([]byte, 1<<16)
	runtime.Stack(buf, true)
	str := strings.Split(string(buf), "\n")[3]

	re := regexp.MustCompile("mocknet\\.\\(\\*[^\\.]*\\.")

	str = re.ReplaceAllLiteralString(str, "")
	str = strings.Split(str, "(")[0]

	time := string(append([]byte(time.Now().Format("2006-01-02T15:04:05.999999999")), []byte{48, 48, 48, 48, 48, 48, 48, 48, 48}...)[:30])

	msg = append([]interface{}{time, str, c.LocalAddr(), c.RemoteAddr()}, msg...)

	format = string(append([]byte("mocknet %30s %s %s->%s "), append([]byte(format), []byte("\n")...)...))

	log.Alertf(format, msg...)
}

func (c *mockConn) Read(b []byte) (n int, err error) {
	return c.Conn.Read(b)
}

func (c *mockConn) Write(b []byte) (n int, err error) {
	if simulationdata.Delay {
		delay := mockDelay(c.LocalAddr().String(), c.RemoteAddr().String())
		time.Sleep(delay)
	}

	n, err = c.Conn.Write(b)

	return
}

func (c *mockConn) Close() error {
	simulationlog.Close(c.LocalAddr().String(), c.RemoteAddr().String())
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

func minNonzeroTime(a, b time.Time) time.Time {
	if a.IsZero() {
		return b
	}
	if b.IsZero() || a.Before(b) {
		return a
	}
	return b
}

//DialTimeout is return Conn
func DialTimeout(networkType, address string, timeout time.Duration, localhost string) (net.Conn, error) {
	connected := make(chan net.Conn)

	now := time.Now()
	earliest := now.Add(timeout)
	ctx, cancel := context.WithDeadline(context.Background(), earliest)
	defer cancel()

	delay := mockDelay(localhost, address)
	go func() {
		time.Sleep(delay)
		connected <- mocknetwork.RegistDial(networkType, address, localhost)
	}()
	select {
	case <-ctx.Done():
		return nil, ErrDialTimeout
	case conn := <-connected:
		mc := &mockConn{
			Conn: conn,
			LocalAddrVal: mockAddr{
				network: networkType,
				address: localhost,
			},
			RemoteAddrVal: mockAddr{
				network: networkType,
				address: address,
			},
			targetID:      address,
			readDeadline:  -1,
			writeDeadline: -1,
		}

		simulationlog.Dial(mc.LocalAddr().String(), mc.RemoteAddr().String(), delay)

		return mc, nil
	}

}

//Dial is return Conn
// func Dial(networkType, address string, localhost string) (net.Conn, error) {
func Dial(networkType, address string, localhost string) (net.Conn, error) {
	// listenAddrs := strings.Split(address, ":")
	// listenPort := listenAddrs[len(listenAddrs)-1]

	// localhost := mocknetwork.GetMainID() + ":" + listenPort

	var conn net.Conn

	delay := mockDelay(localhost, address)
	time.Sleep(delay)

	conn = mocknetwork.RegistDial(networkType, address, localhost)

	mc := &mockConn{
		Conn: conn,
		LocalAddrVal: mockAddr{
			network: networkType,
			address: localhost,
		},
		RemoteAddrVal: mockAddr{
			network: networkType,
			address: address,
		},
		targetID:      address,
		readDeadline:  -1,
		writeDeadline: -1,
	}

	simulationlog.Dial(mc.LocalAddr().String(), mc.RemoteAddr().String(), delay)

	return mc, nil
}

type mockListener struct {
	sync.Mutex
	addr  imockAddr
	node  mocknetwork.NodeInfo
	count int
}

func (l *mockListener) waitAccept() {
	l.node = mocknetwork.RegistAccept(l.addr.String())
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

// Listen announces on the local network address.
func Listen(networkType, addr string) (net.Listener, error) {
	// var addr string
	if strings.Contains(addr, ":") {
		strs := strings.Split(addr, ":")
		if strs[0] == "" {
			addr = mocknetwork.GetMainID() + addr
		}
	}
	var l net.Listener

	ml := mockListener{
		addr: &mockAddr{
			network: networkType,
			address: addr,
		},
	}

	ml.waitAccept()
	simulationlog.Listen(addr)

	l = &ml

	return l, nil
}

/*
type mockPacketConn struct {
	addr         mockAddr
	ls           net.Listener
	readTimeout  time.Time
	writeTimeout time.Time
}

// ReadFrom reads a packet from the connection,
// copying the payload into b. It returns the number of
// bytes copied into b and the return address that
// was on the packet.
// ReadFrom can be made to time out and return
// an Error with Timeout() == true after a fixed time limit;
// see SetDeadline and SetReadDeadline.
func (pc *mockPacketConn) ReadFrom(b []byte) (n int, addr net.Addr, err error) {
	return 0, nil, nil
}

// WriteTo writes a packet with payload b to addr.
// WriteTo can be made to time out and return
// an Error with Timeout() == true after a fixed time limit;
// see SetDeadline and SetWriteDeadline.
// On packet-oriented connections, write timeouts are rare.
func (pc *mockPacketConn) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	return 0, nil
}

// Close closes the connection.
// Any blocked ReadFrom or WriteTo operations will be unblocked and return errors.
func (pc *mockPacketConn) Close() error {
	return nil
}

// LocalAddr returns the local network address.
func (pc *mockPacketConn) LocalAddr() net.Addr {
	return net.Addr(&pc.addr)
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail with a timeout (see type Error) instead of
// blocking. The deadline applies to all future and pending
// I/O, not just the immediately following call to ReadFrom or
// WriteTo. After a deadline has been exceeded, the connection
// can be refreshed by setting a deadline in the future.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful ReadFrom or WriteTo calls.
//
// A zero value for t means I/O operations will not time out.
func (pc *mockPacketConn) SetDeadline(t time.Time) error {
	pc.readTimeout = t
	pc.writeTimeout = t
	return nil
}

// SetReadDeadline sets the deadline for future ReadFrom calls
// and any currently-blocked ReadFrom call.
// A zero value for t means ReadFrom will not time out.
func (pc *mockPacketConn) SetReadDeadline(t time.Time) error {
	pc.readTimeout = t
	return nil
}

// SetWriteDeadline sets the deadline for future WriteTo calls
// and any currently-blocked WriteTo call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means WriteTo will not time out.
func (pc *mockPacketConn) SetWriteDeadline(t time.Time) error {
	pc.writeTimeout = t
	return nil
}

// ListenPacket announces on the local network address.
func ListenPacket(networkType, address string) (net.PacketConn, error) {
	ls, err := Listen(networkType, address)
	if err != nil {
		panic(err)
	}

	pc := mockPacketConn{
		addr: mockAddr{
			network: networkType,
			address: address,
		},
		ls: ls,
	}

	return net.PacketConn(&pc), nil
} */
