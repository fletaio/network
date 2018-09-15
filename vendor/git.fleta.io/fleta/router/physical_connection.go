package router

import (
	"bufio"
	"bytes"
	"net"
	"sync"

	"fleta/util"

	"git.fleta.io/common/log"
	"git.fleta.io/fleta/common"
)

type routerPhysical interface {
	// receiverChan(addr string, genesis Hash256) chan Receiver
	removePhysicalConnenction(pc *physicalConnection) error
	AcceptConn(conn Receiver, genesis common.Coordinate) error
}

//MAGICWORD Start of packet
const MAGICWORD = 'R'

//HANDSHAKE Start of handshake packet
const HANDSHAKE = 'H'

type physicalConnection struct {
	connectionLock sync.Mutex
	writeLock      sync.Mutex

	net.Conn

	addr  RemoteAddr
	lConn map[common.Coordinate]*logicalConnection
	r     routerPhysical
}

func (pc *physicalConnection) run() error {
	for {
		body, genesis, err := pc.readConn()
		// log.Debug("body read len : " + strconv.Itoa(len(body)))
		if err != nil {
			log.Error(err)
			return err
		}
		if body == nil && genesis != nil {
			conn := pc.makeLogicalConnenction(*genesis)
			err := pc.r.AcceptConn(conn, *genesis)
			if err != nil {
				log.Error(err)
			}
		} else {
			go pc.sendToLogicalConn(body, *genesis)
		}
	}
}

func (pc *physicalConnection) sendToLogicalConn(bs []byte, genesis common.Coordinate) (err error) {
	if bs != nil {
		lConn, has := pc.lConn[genesis]
		if !has {
			return ErrNotFoundLogicalConnection
		}
		lConn.Send(bs)
		err = lConn.Flush()
		return
	}
	return nil
}

func (pc *physicalConnection) makeLogicalConnenction(genesis common.Coordinate) Receiver {
	pc.connectionLock.Lock()
	l, has := pc.lConn[genesis]

	if has {
		pc.connectionLock.Unlock()
	} else {
		cChan := make(chan []byte)
		rChan := make(chan []byte)
		rc := &receiver{
			recvChan:   rChan,
			sendChan:   cChan,
			localAddr:  pc.LocalAddr(),
			remoteAddr: pc.RemoteAddr(),
		}
		l = &logicalConnection{
			genesis:           genesis,
			chainSideReceiver: rc,
			Receiver: &receiver{
				recvChan:   cChan,
				sendChan:   rChan,
				localAddr:  pc.LocalAddr(),
				remoteAddr: pc.RemoteAddr(),
			},
		}
		pc.lConn[genesis] = l
		pc.connectionLock.Unlock()

		go pc.runLConn(l)
	}
	return l.chainSideReceiver
}

func (pc *physicalConnection) readConn() (body []byte, genesis *common.Coordinate, returnErr error) {
	genesis = &common.Coordinate{}
	bs, err := pc.readBytes(6 + 5)
	if err != nil {
		return bs, nil, err
	}
	for i, b := range bs[1:7] {
		genesis[i] = b
	}

	if HANDSHAKE == uint8(bs[0]) {
		return nil, genesis, nil
	}

	bodySize := util.BytesToUint32(bs[7:])
	body, err = pc.readBytes(bodySize)
	if err != nil {
		return nil, genesis, err
	}

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

func (pc *physicalConnection) handshake(genesis common.Coordinate) (wrote int64, err error) {
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	if n, err := util.WriteUint8(writer, HANDSHAKE); err == nil {
		wrote += n
	} else {
		return wrote, err
	}

	if n, err := writer.Write(genesis[:]); err == nil {
		wrote += int64(n)
	} else {
		return wrote, err
	}

	if n, err := util.WriteUint32(writer, uint32(0)); err == nil {
		wrote += n
	} else {
		return wrote, err
	}

	writer.Flush()

	if pc.Conn == nil {
		return wrote, ErrNotConnected
	}

	pc.writeLock.Lock()
	if n, err := pc.Conn.Write(b.Bytes()); err == nil {
		pc.writeLock.Unlock()
		wrote = int64(n)
	} else {
		pc.writeLock.Unlock()
		return wrote, err
	}

	return wrote, nil
}

func (pc *physicalConnection) write(body []byte, genesis common.Coordinate) (wrote int64, err error) {
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

	pc.writeLock.Lock()
	if n, err := pc.Conn.Write(b.Bytes()); err == nil {
		pc.writeLock.Unlock()
		wrote = int64(n)
	} else {
		pc.writeLock.Unlock()
		return wrote, err
	}

	// log.Debug("writeConn r : " + pc.Conn.RemoteAddr().String() + " l : " + pc.Conn.LocalAddr().String())
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
		// log.Debug("LogicalConnection Run : " + string(bs))
		if _, err := pc.write(bs, lc.genesis); err != nil {
			return err
		}
	}
}
