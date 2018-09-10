package router

import (
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"git.fleta.io/common/log"
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
		networkType     string
		localhost       string
		remotehost      string
		r               *router
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
				r:               new(),
				chainGenHashStr: "12345678901234567890123456789012",
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
				Conn:  mc,
				lConn: map[Hash256]*logicalConnection{},
				r:     tt.args.r,
			}
			tt.args.r.pConn[RemoteAddr(tt.args.remotehost)] = pConn

			hash, convErr := converterHash256([]byte(tt.args.chainGenHashStr))
			if convErr != nil {
				t.Errorf("New() = %v", convErr)
			}
			ch := tt.args.r.ReceiverChan(tt.args.remotehost, hash)
			pConn.makeLogicalConnenction(hash)

			clientSideReciver := <-ch

			wg := sync.WaitGroup{}
			wg.Add(1)

			var result1 bool

			go func() {
				data, err := clientSideReciver.Recv()
				if err != nil {

				}
				result1 = string(data) == routerSendMsg
				t.Log("clientSide rect : ", string(data), " : ")
				wg.Done()
			}()

			lConn, has := pConn.lConn[hash]
			if !has {
				t.Errorf("cannot found lConn")
			}

			go func() {
				lConn.Send([]byte(routerSendMsg))
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

			hash, convErr := converterHash256([]byte(tt.args.chainGenHashStr))
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
				addr: ":3000",
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

func Test_Dial(t *testing.T) {
	genesis, _ := converterHash256([]byte("12345678901234567890123456789012"))
	type args struct {
		addr    string
		genesis Hash256
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
				addr:    ":3000",
				genesis: genesis,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r1Addr := tt.r1.localAddr(tt.args.addr)

			tt.r1.AddListen(tt.args.addr)
			tt.r2.Dial(r1Addr, tt.args.genesis)

			// listenCh := tt.r1.ReceiverChan(r2Addr, tt.args.genesis)
			dialCh := tt.r2.ReceiverChan(r1Addr, tt.args.genesis)

			recived := false
			select {
			case _, ok := <-dialCh:
				if ok {
					recived = true
				}
			default:
			}

			if tt.want != recived {
				t.Errorf("Dial test wand %v but recived is %v", tt.want, recived)
			}

		})
	}
}

func Test_Dial_Accept(t *testing.T) {
	genesis, _ := converterHash256([]byte("12345678901234567890123456789012"))
	type args struct {
		addr    string
		genesis Hash256
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
				addr:    ":3000",
				genesis: genesis,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r1Addr := tt.r1.localAddr(tt.args.addr)

			tt.r1.AddListen(tt.args.addr)
			tt.r2.Dial(r1Addr, tt.args.genesis)

			listenCh := tt.r1.ReceiverChan(tt.args.addr, tt.args.genesis)
			dialCh := tt.r2.ReceiverChan(tt.args.addr, tt.args.genesis)

			var err error
			dialRecived := false
			select {
			case recv, ok := <-dialCh:
				if ok {
					dialRecived = true
				}
				err = recv.Send([]byte("test send"))
			}

			acceptRecived := false
			select {
			case _, ok := <-listenCh:

				if ok {
					acceptRecived = true
				}
			}

			if err != nil {
				t.Errorf("error detect %v", err)
			}

			if tt.want != dialRecived {
				t.Errorf("Dial test wand %v but dialRecived is %v", tt.want, dialRecived)
			}
			if tt.want != acceptRecived {
				t.Errorf("Dial test wand %v but acceptRecived is %v", tt.want, acceptRecived)
			}

		})
	}
}

func TestRouter_Dial(t *testing.T) {
	genesis, _ := converterHash256([]byte("12345678901234567890123456789012"))
	type args struct {
		addr    string
		genesis Hash256
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
				addr:    ":3000",
				genesis: genesis,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientSendMsg := "send client to router"
			routerSendMsg := "send router to client"
			var result1 bool
			var result2 bool
			r1Addr := tt.r1.localAddr(tt.args.addr)
			r2Addr := tt.r2.localAddr(tt.args.addr)

			tt.r1.AddListen(tt.args.addr)
			tt.r2.Dial(r1Addr, tt.args.genesis)

			ch1 := tt.r1.ReceiverChan(r2Addr, tt.args.genesis)
			ch2 := tt.r2.ReceiverChan(r1Addr, tt.args.genesis)

			wg := sync.WaitGroup{}
			wg.Add(2)

			go func() {
				reciver1 := <-ch1
				go func() {
					// for {
					data, err := reciver1.Recv()
					result1 = string(data) == routerSendMsg
					t.Log("clientSide rect : "+string(data)+" : ", err)
					wg.Done()
					// 	if !ok {
					// 		break
					// 	}
					// }
				}()
				go func() {
					err := reciver1.Send([]byte(clientSendMsg))
					if err != nil && err != io.EOF {
						t.Errorf("client Send err = %v", err)
					}
				}()

			}()
			go func() {
				reciver2 := <-ch2
				go func() {
					// for {
					data, err := reciver2.Recv()
					result2 = string(data) == clientSendMsg
					t.Log("routerSide rect : "+string(data)+" : ", err)
					wg.Done()
					// 	if !ok {
					// 		break
					// 	}
					// }
				}()
				go func() {
					err := reciver2.Send([]byte(routerSendMsg))
					if err != nil && err != io.EOF {
						t.Errorf("router Send err = %v", err)
					}
				}()

			}()

			wg.Wait()

			if (tt.want != result1) || (tt.want != result2) {
				t.Errorf("want %v, but recived router to client is %v and client to router is %v", tt.want, result1, result2)
			}
		})
	}
}

func Test(t *testing.T) {
	genesis, _ := converterHash256([]byte("12345678901234567890123456789012"))
	type args struct {
		addr    string
		genesis Hash256
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
				addr:    ":3000",
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
