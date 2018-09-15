package router

import (
	"io"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"git.fleta.io/common/log"
	"git.fleta.io/fleta/common"
)

type mockAddr struct {
	address string
}

func (c *mockAddr) Network() string { return "tcp" }
func (c *mockAddr) String() string  { return c.address }

type mockConn struct {
	LocalAddrVal  mockAddr
	RemoteAddrVal mockAddr
}

func (c *mockConn) Read(b []byte) (n int, err error)   { log.Debug("Read"); return 0, nil }
func (c *mockConn) Write(b []byte) (n int, err error)  { log.Debug("Write"); return 0, nil }
func (c *mockConn) Close() error                       { log.Debug("Close"); return nil }
func (c *mockConn) LocalAddr() net.Addr                { log.Debug("LocalAddr"); return &c.LocalAddrVal }
func (c *mockConn) RemoteAddr() net.Addr               { log.Debug("RemoteAddr"); return &c.RemoteAddrVal }
func (c *mockConn) SetDeadline(t time.Time) error      { log.Debug("SetDeadline"); return nil }
func (c *mockConn) SetReadDeadline(t time.Time) error  { log.Debug("SetReadDeadline"); return nil }
func (c *mockConn) SetWriteDeadline(t time.Time) error { log.Debug("SetWriteDeadline"); return nil }

func TestClientAndRouterCommunicate(t *testing.T) {
	type args struct {
		networkType        string
		localhost          string
		remotehost         string
		r                  *router
		chainGenCoordinate common.Coordinate
	}
	tests := []struct {
		name string
		args
		want bool
	}{
		{
			name: "test",
			args: args{
				networkType:        "tcp",
				localhost:          ":3000",
				remotehost:         "test2:3000",
				r:                  new(),
				chainGenCoordinate: common.Coordinate{},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// clientSendMsg := "send client to router"
			routerSendMsg := "send router to client"

			mc := &mockConn{
				LocalAddrVal: mockAddr{
					address: tt.args.localhost,
				},
				RemoteAddrVal: mockAddr{
					address: tt.args.remotehost,
				},
			}

			pConn := &physicalConnection{
				addr:  RemoteAddr(tt.args.remotehost),
				Conn:  mc,
				lConn: map[common.Coordinate]*logicalConnection{},
				r:     tt.args.r,
			}
			tt.args.r.pConn[RemoteAddr(tt.args.remotehost)] = pConn

			chainSideConn := pConn.makeLogicalConnenction(tt.args.chainGenCoordinate)

			wg := sync.WaitGroup{}
			wg.Add(1)

			var result1 bool

			go func() {
				data, err := chainSideConn.Recv()
				if err != nil {

				}
				result1 = string(data) == routerSendMsg
				t.Log("clientSide rect : ", string(data), " : ")
				wg.Done()
			}()

			lConn, has := pConn.lConn[tt.args.chainGenCoordinate]
			if !has {
				t.Errorf("cannot found lConn")
			}

			go func() {
				lConn.Send([]byte(routerSendMsg))
				lConn.Flush()
			}()

			wg.Wait()

			if tt.want != result1 {
				t.Errorf("want %v, but recived router to client is %v", tt.want, result1)
			}
		})
	}
}

/* func TestClosedClientCommunicate(t *testing.T) {
	type args struct {
		networkType     string
		localhost       string
		remotehost      string
		r               *Router
		chainGenHashStr string
	}
	tests := []struct {
		name string
		args
		want bool
	}{
		{
			name: "test",
			args: args{
				networkType:     "tcp",
				localhost:       "test1:3000",
				remotehost:      "test2:3000",
				r:               New(),
				chainGenHashStr: "12345678901234567890123456789012",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientSendMsg := "send client to router"
			routerSendMsg := "send router to client"

			mc := &mockConn{
				LocalAddrVal: mockAddr{
					address: tt.args.localhost,
				},
				RemoteAddrVal: mockAddr{
					address: tt.args.remotehost,
				},
			}

			pConn := &PhysicalConnection{
				Conn:  mc,
				lConn: map[Hash256]*LogicalConnection{},
				r:     tt.args.r,
				// ReceiverChanMap: map[Hash256]chan Receiver{},
			}
			tt.args.r.pConn[RemoteAddr(tt.args.remotehost)] = pConn

			hash, convErr := ConverterHash256([]byte(tt.args.chainGenHashStr))
			if convErr != nil {
				t.Errorf("New() = %v", convErr)
			}
			ch := tt.args.r.ReceiverChan(tt.args.remotehost, hash)
			pConn.makeLogicalConnenction(hash)

			clientSideReciver := <-ch

			wg := sync.WaitGroup{}
			wg.Add(2)

			var result1 bool
			var result2 bool

			lConn, has := pConn.lConn[hash]
			if !has {
				t.Errorf("cannot found lConn")
			}

			go func() {
				for {
					data, err := clientSideReciver.Recv()
					result1 = string(data) == routerSendMsg
					t.Log("clientSide rect : " + string(data) + " : " + err.Error())
					if err != nil {
						break
					}
				}
			}()
			go func() {
				err := clientSideReciver.Send([]byte(clientSendMsg))
				if err != nil && err != io.EOF {
					t.Errorf("client Send err = %v", err)
				}
			}()

			lConn, has = pConn.lConn[hash]
			if !has {
				t.Errorf("cannot found lConn")
			}

			clientSideReciver.Close()
			time.Sleep(time.Second)

			go func() {
				for {
					data, err := lConn.Recv()
					result2 = string(data) == clientSendMsg
					t.Log("routerSide rect : " + string(data) + " : " + err.Error())
					if err != nil {
						break
					}
				}
			}()
			go func() {
				err := lConn.Send([]byte(routerSendMsg))
				if err != nil && err != io.EOF {
					t.Errorf("router Send err = %v", err)
				}
			}()

			time.Sleep(time.Second * 2)
			if (tt.want != result1) || (tt.want != result2) {
				t.Errorf("want %v, but recived router to client is %v and client to router is %v", tt.want, result1, result2)
			}
		})
	}
} */

func TestRouter_AddListen(t *testing.T) {
	type args struct {
		addr string
	}
	tests := []struct {
		name string
		r    *router
		args args
		want int
	}{
		{
			name: "test",
			r:    new(),
			args: args{
				addr: ":3002",
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.r.AddListen(tt.args.addr)
			tt.r.AddListen(tt.args.addr)
			tt.r.AddListen(tt.args.addr)

			if len(tt.r.Listeners) != tt.want {
				t.Errorf("tt.r.Listeners length want = %v, but length is %v", tt.want, len(tt.r.Listeners))
			}

		})
	}
}

func Test_Dial_Accept(t *testing.T) {
	var genesis common.Coordinate
	copy(genesis[:], []byte("123451"))
	type args struct {
		addr    string
		genesis common.Coordinate
	}
	tests := []struct {
		name string
		r1   *router
		r2   *router
		args args
		want bool
	}{
		{
			name: "test",
			r1:   new(),
			r2:   new(),
			args: args{
				addr:    "test:3003",
				genesis: genesis,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.r1.AddListen(tt.args.addr)

			wg := sync.WaitGroup{}
			wg.Add(2)

			var result1 bool

			go func() {
				conn, _ := tt.r2.Dial(tt.args.addr, tt.args.genesis)
				conn.Send([]byte("result"))
				conn.Flush()
				wg.Done()
			}()
			go func() {
				conn, _ := tt.r1.Accept(tt.args.addr, tt.args.genesis)
				bs, _ := conn.Recv()
				result1 = string(bs) == "result"
				wg.Done()
			}()

			wg.Wait()

			if tt.want != result1 {
				t.Errorf("Dial test wand %v but recived is %v", tt.want, result1)
			}

		})
	}
}

func TestRouter_Data_PingPong(t *testing.T) {
	var genesis1 common.Coordinate
	copy(genesis1[:], []byte("123451"))
	type args struct {
		addr1    string
		addr2    string
		genesis1 common.Coordinate
	}
	tests := []struct {
		name    string
		r1      *router
		r2      *router
		r3      *router
		args    args
		wantErr error
	}{
		{
			name: "test",
			r1:   new(),
			r2:   new(),
			r3:   new(),
			args: args{
				addr1:    "test:3005",
				addr2:    "test:3006",
				genesis1: genesis1,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.r1.AddListen(tt.args.addr1)
			// tt.r2.AddListen(tt.args.addr2)

			wg := sync.WaitGroup{}
			wg.Add(4)

			var resultErr error

			go func() {
				for {
					conn, _ := tt.r1.Accept(tt.args.addr1, tt.args.genesis1)
					err := read(conn, "ch1 1")
					if err != nil {
						resultErr = err
					}
					wg.Done()
				}
			}()
			go func() {
				conn, _ := tt.r2.Dial(tt.args.addr1, tt.args.genesis1)
				err := read(conn, "ch1 2")
				if err != nil {
					resultErr = err
				}
				wg.Done()
			}()
			go func() {
				time.Sleep(time.Millisecond * 100)
				conn, _ := tt.r3.Dial(tt.args.addr1, tt.args.genesis1)
				err := read(conn, "ch1 3")
				if err != nil {
					resultErr = err
				}
				wg.Done()
			}()

			wg.Wait()

			if tt.wantErr != resultErr {
				t.Errorf("Dial success and expect err is %v but return error is %v", tt.wantErr, resultErr)
			}

		})
	}
}

func read(reciver Receiver, startStr string) error {
	reciver.Send([]byte(startStr))
	reciver.Flush()

	for {
		data, err := reciver.Recv()
		log.Debug(string(data))
		if strings.Contains(string(data), "rand") {
			continue
		}
		str := string(data)
		strs := strings.Split(str, ":")
		i, err := strconv.Atoi(strs[len(strs)-1])
		if err != nil {
			strs = append(strs, "1")
		} else {
			i++
			strs[len(strs)-1] = strconv.Itoa(i)
		}
		str = strings.Join(strs, ":")

		time.Sleep(time.Millisecond * time.Duration(rand.Intn(10)))

		reciver.Send([]byte(str))
		err = reciver.Flush()
		if err != nil {
			if err != io.EOF {
				log.Debug(string(data)+" : ", err)
				break
			}
			return err
		}
		if i > 5 {
			return nil
		}
	}
	return nil
}

func TestRouter_Data_PingPong_multy_chain(t *testing.T) {
	var genesis1 common.Coordinate
	var genesis2 common.Coordinate
	var genesis3 common.Coordinate
	var genesis4 common.Coordinate
	copy(genesis1[:], []byte("123451"))
	copy(genesis2[:], []byte("123452"))
	copy(genesis3[:], []byte("123453"))
	copy(genesis4[:], []byte("123454"))
	type args struct {
		addr1    string
		addr2    string
		genesis1 common.Coordinate
		genesis2 common.Coordinate
		genesis3 common.Coordinate
		genesis4 common.Coordinate
	}
	tests := []struct {
		name string
		r1   *router
		r2   *router
		args args
		want int
	}{
		{
			name: "test",
			r1:   new(),
			r2:   new(),
			args: args{
				addr1:    "test:3005",
				addr2:    "test:3006",
				genesis1: genesis1,
				genesis2: genesis2,
				genesis3: genesis3,
				genesis4: genesis4,
			},
			want: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.r1.AddListen(tt.args.addr1)
			// tt.r2.AddListen(tt.args.addr2)

			wg := sync.WaitGroup{}
			wg.Add(8)

			go func() {
				conn, _ := tt.r1.Accept(tt.args.addr1, tt.args.genesis1)
				read(conn, "ch1 1")
				wg.Done()
			}()
			go func() {
				conn, _ := tt.r2.Dial(tt.args.addr1, tt.args.genesis1)
				read(conn, "ch1 2")
				wg.Done()
			}()

			go func() {
				time.Sleep(time.Millisecond * 1000)
				go func() {
					conn, _ := tt.r1.Accept(tt.args.addr1, tt.args.genesis2)
					read(conn, "ch2 1")
					wg.Done()
				}()
				go func() {
					conn, _ := tt.r2.Dial(tt.args.addr1, tt.args.genesis2)
					read(conn, "ch2 2")
					wg.Done()
				}()
			}()
			go func() {
				conn, _ := tt.r1.Accept(tt.args.addr1, tt.args.genesis3)
				read(conn, "ch3 1")
				wg.Done()
			}()
			go func() {
				conn, _ := tt.r2.Dial(tt.args.addr1, tt.args.genesis3)
				read(conn, "ch3 2")
				wg.Done()
			}()
			go func() {
				time.Sleep(time.Millisecond * 1000)
				go func() {
					conn, _ := tt.r1.Accept(tt.args.addr1, tt.args.genesis4)
					read(conn, "ch4 1")
					wg.Done()
				}()
				go func() {
					conn, _ := tt.r2.Dial(tt.args.addr1, tt.args.genesis4)
					read(conn, "ch4 2")
					wg.Done()
				}()
			}()

			// time.Sleep(time.Second)
			wg.Wait()
		})
	}
}

func Test(t *testing.T) {
	var genesis common.Coordinate
	copy(genesis[:], []byte("123451"))
	type args struct {
		addr    string
		genesis common.Coordinate
	}
	tests := []struct {
		name string
		r1   *router
		r2   *router
		args args
		want bool
	}{
		{
			name: "test",
			r1:   new(),
			r2:   new(),
			args: args{
				addr:    ":3007",
				genesis: genesis,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}
