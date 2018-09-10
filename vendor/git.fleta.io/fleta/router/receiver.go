package router

import "io"

type logicalConnection struct {
	genesis Hash256
	*Receiver
}

//Receiver is data communication unit
type Receiver struct {
	recvChan <-chan []byte
	sendChan chan<- []byte
	isClosed bool
}

//Recv is receive
func (r *Receiver) Recv() ([]byte, error) {
	data, ok := <-r.recvChan
	if !ok {
		r.Close()
		return nil, io.EOF
	}
	return data, nil
}

//Send is send
func (r *Receiver) Send(data []byte) (err error) {
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

	r.sendChan <- data

	return err
}

//Close is data communication channel close
func (r *Receiver) Close() {
	if !r.isClosed {
		r.isClosed = true
		close(r.sendChan)
	}
}
