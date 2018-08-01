package formulator

import (
	"errors"
	"fleta/flanetinterface"
	"fleta/util"
	"io"
	"time"
)

//formulator object err list
var (
	ErrNotEnoughLength = errors.New("ErrNotEnoughLength")
)

//Node is formulator node struct
type Node struct {
	Address    string
	arrayIndex uint64
	Detected   time.Time
	Block      time.Time
}

//Addr TODO
func (n *Node) Addr() string {
	return n.Address
}

//Type TODO
func (n *Node) Type() string {
	return flanetinterface.FormulatorNode
}

// DetectedTime TODO
func (n *Node) DetectedTime() time.Time {
	return n.Detected
}

// BlockTime TODO
func (n *Node) BlockTime() time.Time {
	return n.Block
}

func writeTo(w io.Writer, bda ...[]byte) (int64, error) {
	var wrote int64

	for _, ba := range bda {
		num, err := util.WriteUint16(w, uint16(len(ba)))
		if err != nil {
			return wrote, err
		}
		wrote += num

		n, err := w.Write(ba)
		if err != nil {
			return wrote, err
		}
		wrote += int64(n)
	}

	return wrote, nil
}

// WriteTo TODO
func (n *Node) WriteTo(w io.Writer) (int64, error) {
	var wrote int64

	addrba := []byte(n.Address)
	dba, err := n.Detected.MarshalBinary()
	if err != nil {
		return wrote, err
	}
	bba, err := n.Block.MarshalBinary()
	if err != nil {
		return wrote, err
	}

	return writeTo(w, addrba, dba, bba)
	// return writeTo(w, addrba)
}

// ReadFrom TODO
func readFrom(r io.Reader) ([][]byte, int64, error) {
	var read int64
	result := make([][]byte, 0)

	for {
		Len, n64, err := util.ReadUint16(r)
		if err != nil {
			return result, read, err
		}
		read += n64
		bs := make([]byte, Len)
		n, err := r.Read(bs)
		if err != nil {
			return result, read, err
		}
		read += int64(n)
		result = append(result, bs)
	}
}

// ReadFrom TODO
func (n *Node) ReadFrom(r io.Reader) (int64, error) {
	result, read, _ := readFrom(r)
	if len(result) != 3 {
		return read, ErrNotEnoughLength
	}

	n.Address = string(result[0])
	err := n.Detected.UnmarshalBinary(result[1])
	if err != nil {
		return read, err
	}
	err = n.Block.UnmarshalBinary(result[2])
	if err != nil {
		return read, err
	}

	return read, nil
}
