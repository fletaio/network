package mocknet

import (
	"net"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"git.fleta.io/fleta/mocknet/string_util"

	"git.fleta.io/fleta/framework/log"
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
	for i := 1; i < length; i++ {
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
	targetID      string
	readDeadline  time.Duration
	writeDeadline time.Duration
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

	log.Infof(format, msg...)
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
