package router

import (
	"net"
	"strconv"
	"strings"
	"sync"

	"fleta/samutil"

	"git.fleta.io/common/log"
	"git.fleta.io/fleta/common"
)

//RemoteAddr is remote address type
type RemoteAddr string

//Router router interface
type Router interface {
	AddListen(addr string) error
	Dial(addrStr string, genesis common.Coordinate) (Receiver, error)
	Accept(addrStr string, genesis common.Coordinate) (Receiver, error)
}

type router struct {
	receiverLock    sync.Mutex
	pConnLock       sync.Mutex
	acceptLock      sync.Mutex
	Listeners       map[RemoteAddr]net.Listener
	pConn           map[RemoteAddr]*physicalConnection
	ReceiverChanMap map[int]map[common.Coordinate]chan Receiver
	RouterID        string
}

var i = 0

//New is creator of router
func New() Router {
	return new()
}

func new() *router {
	r := &router{
		Listeners:       map[RemoteAddr]net.Listener{},
		pConn:           map[RemoteAddr]*physicalConnection{},
		ReceiverChanMap: map[int]map[common.Coordinate]chan Receiver{},
		RouterID:        samutil.Sha256HexInt(i),
	}
	i++
	return r
}

func (r *router) localAddr(addr string) string {
	ss := strings.Split(addr, ":")
	port := ss[len(ss)-1]
	return r.RouterID + ":" + port
}

func (r *router) AddListen(addr string) error {
	_, has := r.Listeners[RemoteAddr(addr)]
	if !has {
		// l, err := mocknet.Listen("tcp", addr)
		l, err := net.Listen("tcp", addr)
		log.Debug("Listen ", addr, l.Addr().String())
		if err != nil {
			return err
		}
		r.Listeners[RemoteAddr(addr)] = l
		go r.run(l)
	}
	return nil
}

func (r *router) run(l net.Listener) {
	for {
		conn, err := l.Accept()
		log.Debug("Router Run Accept " + conn.LocalAddr().String() + " : " + conn.RemoteAddr().String())
		if err != nil {
			log.Error(err)
			continue
		}
		r.pConnLock.Lock()
		addr := RemoteAddr(conn.RemoteAddr().String())
		_, has := r.pConn[addr]
		if has {
			conn.Close()
		} else {
			pc := &physicalConnection{
				addr:  addr,
				Conn:  conn,
				lConn: map[common.Coordinate]*logicalConnection{},
				r:     r,
			}
			r.pConn[addr] = pc
			go pc.run()
		}
		r.pConnLock.Unlock()
	}
}

func (r *router) Accept(addrStr string, genesis common.Coordinate) (Receiver, error) {
	r.acceptLock.Lock()
	l, has := r.Listeners[RemoteAddr(addrStr)]
	if !has {
		r.acceptLock.Unlock()
		return nil, ErrListenFirst
	}
	lAddr := strings.Split(l.Addr().String(), ":")
	portStr := lAddr[len(lAddr)-1]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		r.acceptLock.Unlock()
		return nil, err
	}

	hashMap, has := r.ReceiverChanMap[port]
	if !has {
		hashMap = map[common.Coordinate]chan Receiver{}
		r.ReceiverChanMap[port] = hashMap
	}

	ch, has := hashMap[genesis]
	if !has {
		ch = make(chan Receiver)
		hashMap[genesis] = ch
	}
	r.acceptLock.Unlock()

	return <-ch, nil
}

func (r *router) AcceptConn(conn Receiver, genesis common.Coordinate) error {
	r.acceptLock.Lock()
	listenAddrs := strings.Split(conn.LocalAddr().String(), ":")
	listenAddr := ":" + listenAddrs[len(listenAddrs)-1]
	l, has := r.Listeners[RemoteAddr(listenAddr)]
	if !has {
		r.acceptLock.Unlock()
		return ErrListenFirst
	}
	lAddr := strings.Split(l.Addr().String(), ":")
	portStr := lAddr[len(lAddr)-1]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		r.acceptLock.Unlock()
		return err
	}

	hashMap, has := r.ReceiverChanMap[port]
	if !has {
		hashMap = map[common.Coordinate]chan Receiver{}
		r.ReceiverChanMap[port] = hashMap
	}

	ch, has := hashMap[genesis]
	if !has {
		ch = make(chan Receiver)
		hashMap[genesis] = ch
	}
	r.acceptLock.Unlock()

	ch <- conn
	delete(hashMap, genesis)
	return nil
}

func (r *router) Dial(addrStr string, genesis common.Coordinate) (Receiver, error) {
	r.pConnLock.Lock()
	defer r.pConnLock.Unlock()
	addr := RemoteAddr(addrStr)
	pConn, has := r.pConn[addr]
	if !has {
		localhost := r.localAddr(addrStr)
		log.Debug("Dial ", " ", addrStr, " ", localhost)
		// conn, err := mocknet.Dial("tcp", addrStr, localhost)
		conn, err := net.Dial("tcp", addrStr)
		if err != nil {
			return nil, err
		}
		pConn = &physicalConnection{
			addr:  addr,
			Conn:  conn,
			lConn: map[common.Coordinate]*logicalConnection{},
			r:     r,
		}
		r.pConn[addr] = pConn
		go pConn.run()
	}
	pConn.handshake(genesis)

	return pConn.makeLogicalConnenction(genesis), nil
}

func (r *router) removePhysicalConnenction(pc *physicalConnection) error {
	r.pConnLock.Lock()
	delete(r.pConn, pc.addr)
	r.pConnLock.Unlock()
	return pc.Close()
}

// func (r *router) receiverChan(addr string, genesis Hash256) chan Receiver {
// 	strs := strings.Split(addr, ":")
// 	stringPort := strs[len(strs)-1]
// 	port, err := strconv.Atoi(stringPort)
// 	if err != nil {
// 		return nil
// 	}

// 	r.receiverLock.Lock()
// 	portLv, has := r.ReceiverChanMap[port]
// 	if !has {
// 		portLv = make(map[Hash256]chan Receiver)
// 		r.ReceiverChanMap[port] = portLv
// 	}

// 	ch, has := portLv[genesis]
// 	if !has {
// 		ch = make(chan Receiver, 1024)
// 		portLv[genesis] = ch
// 	}
// 	r.receiverLock.Unlock()
// 	return ch
// }

// func (r *router) ReceiverChan(addr string, genesis Hash256) <-chan Receiver {
// 	return r.receiverChan(addr, genesis)
// }
