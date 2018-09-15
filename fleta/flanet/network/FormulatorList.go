package network

import (
	"encoding/json"
	"io"
	"log"

	"fleta/message"
	"fleta/util"
)

type FormulatorList struct {
	message.Message
	List []string
}

func NewFormulatorList(list []string) *FormulatorList {
	return &FormulatorList{
		List: list,
	}
}

func (rf *FormulatorList) WriteTo(w io.Writer) (int64, error) {
	var wrote int64

	marshal, err := json.Marshal(rf)
	if err != nil {
		log.Fatal("Cannot encode to JSON ", err)
	}

	num, err := util.WriteUint16(w, uint16(len(marshal)))
	if err != nil {
		return wrote, err
	}
	wrote += num

	n, err := w.Write(marshal)
	if err != nil {
		return wrote, err
	}
	wrote += int64(n)

	return wrote, nil
}

func (rf *FormulatorList) ReadFrom(r io.Reader) (int64, error) {
	var read int64

	Len, n64, err := util.ReadUint16(r)
	if err != nil {
		return read, err
	}
	read += n64
	bs := make([]byte, Len)
	n, err := r.Read(bs)
	if err != nil {
		return read, err
	}
	read += int64(n)

	json.Unmarshal(bs, &rf)
	return read, nil

	// rf.List[count] = string(bs)

	// var read int64
	// Len, n64, err := util.ReadUint16(r)
	// if err != nil {
	// 	return read, err
	// }
	// read += n64

	// rf.List = make([]string, Len)

	// var count int
	// for {
	// 	Len, n64, err := util.ReadUint16(r)
	// 	if err != nil {
	// 		return read, err
	// 	}
	// 	read += n64
	// 	bs := make([]byte, Len)
	// 	n, err := r.Read(bs)
	// 	if err != nil {
	// 		return read, err
	// 	}
	// 	read += int64(n)
	// 	rf.List[count] = string(bs)
	// 	count++
	// }

}
