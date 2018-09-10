package router

import (
	"net"
	"strconv"
	"strings"

	"fleta/mock/mocknet"
	"fleta/samutil"

	"git.fleta.io/common/log"
)

//Hash256 is genesis hash type
type Hash256 [32]byte

//RemoteAddr is remote address type
type RemoteAddr string

//Router router interface
type Router interface {
	AddListen(addr string) error
	Dial(addrStr string, genesis Hash256) error
	ReceiverChan(addr string, genesis Hash256) <-chan Receiver
}

type router struct {
	Listeners       map[RemoteAddr]net.Listener
	pConn           map[RemoteAddr]*physicalConnection
	ReceiverChanMap map[int]map[Hash256]chan Receiver
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
		ReceiverChanMap: map[int]map[Hash256]chan Receiver{},
		RouterID:        samutil.Sha256HexInt(i),
	}
	i++
	return r
}

func converterHash256(in []byte) (out Hash256, err error) {
	if len(in) != 32 {
		err = ErrMismatchHashSize
		return
	}

	copy(out[:], in)
	return
}

func (r *router) localAddr(addr string) string {
	ss := strings.Split(addr, ":")
	port := ss[len(ss)-1]
	return r.RouterID + ":" + port
}

func (r *router) AddListen(addr string) error {
	_, has := r.Listeners[RemoteAddr(addr)]
	if !has {
		l, err := mocknet.Listen("tcp", r.localAddr(addr))
		log.Debug("Listen " + r.localAddr(addr))
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
		addr := RemoteAddr(conn.RemoteAddr().String())
		pc := &physicalConnection{
			Conn:  conn,
			lConn: map[Hash256]*logicalConnection{},
			r:     r,
		}
		r.pConn[addr] = pc
		go pc.run()
	}
}

func (r *router) Dial(addrStr string, genesis Hash256) error {
	addr := RemoteAddr(addrStr)
	pConn, has := r.pConn[addr]
	if !has {
		conn, err := mocknet.Dial("tcp", addrStr, r.localAddr(addrStr))
		log.Debug("Dial to " + addrStr + " from " + r.localAddr(addrStr))
		if err != nil {
			return err
		}
		pConn = &physicalConnection{
			Conn:  conn,
			lConn: map[Hash256]*logicalConnection{},
			r:     r,
		}
		r.pConn[addr] = pConn
		go pConn.run()
	}

	pConn.makeLogicalConnenction(genesis)
	return nil
}

func (r *router) removePhysicalConnenction(pc *physicalConnection) error {
	delete(r.pConn, RemoteAddr(pc.RemoteAddr().String()))
	return pc.Close()
}

func (r *router) receiverChan(addr string, genesis Hash256) chan Receiver {
	strs := strings.Split(addr, ":")
	stringPort := strs[len(strs)-1]
	port, err := strconv.Atoi(stringPort)
	if err != nil {
		return nil
	}

	portLv, has := r.ReceiverChanMap[port]
	if !has {
		portLv = make(map[Hash256]chan Receiver)
		r.ReceiverChanMap[port] = portLv
	}

	ch, has := portLv[genesis]
	if !has {
		ch = make(chan Receiver, 1000)
		portLv[genesis] = ch
	}
	return ch
}

func (r *router) ReceiverChan(addr string, genesis Hash256) <-chan Receiver {
	return r.receiverChan(addr, genesis)
}
