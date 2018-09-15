package router

import (
	"bytes"
	"io"
	"net"

	"git.fleta.io/fleta/common"
)

type logicalConnection struct {
	genesis           common.Coordinate
	chainSideReceiver Receiver
	Receiver
}

//Receiver is data communication unit
type Receiver interface {
	Recv() ([]byte, error)
	Write(data []byte) (int, error)
	Send(data []byte) error
	Flush() error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	Close()
}

//Receiver is data communication unit
type receiver struct {
	recvChan   <-chan []byte
	sendChan   chan<- []byte
	b          *bytes.Buffer
	localAddr  net.Addr
	remoteAddr net.Addr
	isClosed   bool
}

//Recv is receive
func (r *receiver) Recv() ([]byte, error) {
	data, ok := <-r.recvChan
	if !ok {
		r.Close()
		return nil, io.EOF
	}
	return data, nil
}

//Send is send
func (r *receiver) Write(data []byte) (int, error) {
	if r.b == nil {
		r.b = &bytes.Buffer{}
	}
	return r.b.Write(data)
}

//Send is send
func (r *receiver) Send(data []byte) (err error) {
	if r.b == nil {
		r.b = &bytes.Buffer{}
	}
	_, err = r.b.Write(data)
	return
}

//Flush is flush
func (r *receiver) Flush() (err error) {
	if r.isClosed {
		return io.EOF
	}
	defer func() {
		if rc := recover(); rc != nil {
			if _, is := rc.(error); is {
				err = io.EOF
			}
		}
	}()
	if r.b == nil {
		r.sendChan <- []byte{}
		return
	}

	b := r.b.Bytes()
	r.b = &bytes.Buffer{}
	r.sendChan <- b

	return
}

//LocalAddr is local address infomation
func (r *receiver) LocalAddr() net.Addr {
	return r.localAddr
}

//RemoteAddr is remote address infomation
func (r *receiver) RemoteAddr() net.Addr {
	return r.remoteAddr
}

//Close is close the data communicate channel
func (r *receiver) Close() {
	if !r.isClosed {
		r.isClosed = true
		close(r.sendChan)
	}
}
