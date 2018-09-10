package router

import (
	"bufio"
	"bytes"
	"net"
	"sync"

	"git.fleta.io/common/log"
	"git.fleta.io/fleta/util"
)

type routerPhysical interface {
	receiverChan(addr string, genesis Hash256) chan Receiver
	removePhysicalConnenction(pc *physicalConnection) error
}

//MAGICWORD Start of packet
const MAGICWORD = 'R'

type physicalConnection struct {
	sync.Mutex
	net.Conn
	lConn map[Hash256]*logicalConnection
	r     routerPhysical
}

func (pc *physicalConnection) run() error {
	for {
		body, genesis, err := pc.readConn()
		if err != nil {
			log.Error(err)
			return err
		}
		pc.sendToLogicalConn(body, *genesis)
	}
}

func (pc *physicalConnection) sendToLogicalConn(bs []byte, genesis Hash256) (err error) {
	pc.makeLogicalConnenction(genesis)

	lConn, has := pc.lConn[genesis]
	if !has {
		return ErrNotFoundLogicalConnection
	}
	return lConn.Send(bs)
}

func (pc *physicalConnection) makeLogicalConnenction(genesis Hash256) *Receiver {
	l, has := pc.lConn[genesis]
	if !has {
		cChan := make(chan []byte)
		rChan := make(chan []byte)
		rc := &Receiver{
			recvChan: rChan,
			sendChan: cChan,
		}
		l = &logicalConnection{
			genesis,
			&Receiver{
				recvChan: cChan,
				sendChan: rChan,
			},
		}
		pc.lConn[genesis] = l

		go pc.runLConn(l)
		if pc.r == nil {
			return rc
		}
		ch := pc.r.receiverChan(pc.LocalAddr().String(), genesis)
		ch <- *rc
	}
	return nil
}

func (pc *physicalConnection) readConn() (body []byte, genesis *Hash256, returnErr error) {
	bs, err := pc.readBytes(37)
	if err != nil {
		return bs, nil, err
	}

	bodySize := util.BytesToUint32(bs[33:])
	body, err = pc.readBytes(bodySize)
	if err != nil {
		return nil, nil, err
	}

	_genesis, err := converterHash256(bs[1:33])
	if err != nil {
		return nil, nil, err
	}
	genesis = &_genesis

	return
}

func (pc *physicalConnection) readBytes(n uint32) (read []byte, returnErr error) {
	readedN := uint32(0)
	for readedN < n {
		bs := make([]byte, n-readedN)
		readN, err := pc.Conn.Read(bs)
		if err != nil {
			return read, err
		}
		readedN += uint32(readN)
		read = append(read, bs[:readN]...)
	}
	return
}

func (pc *physicalConnection) write(body []byte, genesis Hash256) (wrote int64, err error) {
	len := len(body)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	if n, err := util.WriteUint8(writer, MAGICWORD); err == nil {
		wrote += n
	} else {
		return wrote, err
	}

	if n, err := writer.Write(genesis[:]); err == nil {
		wrote += int64(n)
	} else {
		return wrote, err
	}

	if n, err := util.WriteUint32(writer, uint32(len)); err == nil {
		wrote += n
	} else {
		return wrote, err
	}

	if n, err := writer.Write(body); err == nil {
		wrote += int64(n)
	} else {
		return wrote, err
	}

	writer.Flush()

	if pc.Conn == nil {
		return wrote, ErrNotConnected
	}

	if n, err := pc.Conn.Write(b.Bytes()); err == nil {
		wrote = int64(n)
	} else {
		return wrote, err
	}

	log.Debug("writeConn r : " + pc.Conn.RemoteAddr().String() + " l : " + pc.Conn.LocalAddr().String())
	return wrote, nil
}

func (pc *physicalConnection) runLConn(lc *logicalConnection) error {
	defer func() {
		delete(pc.lConn, lc.genesis)
		lc.Close()
		if len(pc.lConn) == 0 {
			pc.r.removePhysicalConnenction(pc)
		}
	}()
	for {
		bs, err := lc.Recv()
		if err != nil {
			return err
		}
		log.Debug("LogicalConnection Run : " + string(bs))
		if _, err := pc.write(bs, lc.genesis); err != nil {
			return err
		}
	}
}
