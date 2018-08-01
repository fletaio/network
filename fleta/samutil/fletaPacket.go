package samutil

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
)

//MagicWord is FletaPacket's magicword
const MagicWord = "F"

// packet header size
const (
	SizeOfMagicWord   = 1
	SizeOfCommand     = 8
	SizeOfCompression = 1
	SizeOfSize        = 4
)

// packet header index
const (
	IndexOfMagicWord   = 0
	IndexOfCommand     = SizeOfMagicWord
	IndexOfCompression = IndexOfCommand + SizeOfCommand
	IndexOfSize        = IndexOfCompression + SizeOfCompression
	IndexOfContent     = IndexOfSize + SizeOfSize
)

//MagicWordByte is FletaPacket's MagicWordByte
var MagicWordByte []byte

func init() {
	MagicWordByte = []byte(MagicWord)
}

//IsStartMagicWord is check the packet start with magic word
func IsStartMagicWord(tmp []byte) bool {
	for i, by := range MagicWordByte {
		if tmp[i] != by {
			return false
		}
	}
	return true
	// if tmp[0] == MagicWordByte[0] && tmp[1] == MagicWordByte[1] && tmp[2] == MagicWordByte[2] && tmp[3] == MagicWordByte[3] {
	// 	return true
	// }
	// return false
}

//FletaPacket struct
//MagicWord(8), Command(64), Compression(8), Size(32), Content
type FletaPacket struct {
	Command     string
	Compression bool
	Content     string
}

// ErrPacket list
var (
	ErrInvalidPacket            = errors.New("Invalid packet")
	ErrInvalidLengthOfMagicWord = errors.New("Invalid magicWord length")
	ErrInvalidLengthOfCommand   = errors.New("Invalid command length")
)

func (fp *FletaPacket) validate() error {
	if len(MagicWord) != SizeOfMagicWord {
		return ErrInvalidLengthOfMagicWord
	}
	if len(fp.Command) != SizeOfCommand {
		return ErrInvalidLengthOfCommand
	}
	return nil
}

func compress(src string) ([]byte, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write([]byte(src)); err != nil {
		return nil, err
	}
	if err := gz.Flush(); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

//Packet is struct to packet
func (fp *FletaPacket) Packet() ([]byte, error) {
	if err := fp.validate(); err != nil {
		return nil, err
	}

	var packet []byte
	packet = append(packet, []byte(MagicWord)...)
	packet = append(packet, []byte(fp.Command)...)

	var content []byte
	if fp.Compression {
		packet = append(packet, 1)
		con, err := compress(fp.Content)
		if err != nil {
			return nil, err
		}
		content = []byte(con)
	} else {
		packet = append(packet, 0)
		content = []byte(fp.Content)
	}

	var size uint32
	size = uint32(len(content))
	sizeByte := make([]byte, SizeOfSize)
	binary.LittleEndian.PutUint32(sizeByte, size)
	packet = append(packet, sizeByte...)
	packet = append(packet, content...)

	return packet, nil
}

//FletaPacketToStruct is marshal packet
func FletaPacketToStruct(buf []byte) FletaPacket {
	commend := string(buf[IndexOfCommand : IndexOfCommand+SizeOfCommand])
	var compression bool
	var content string
	if buf[IndexOfCompression] == 1 {
		compression = true
		// data, _ := base64.StdEncoding.DecodeString(string(buf[17:]))
		rdata := bytes.NewReader(buf[IndexOfContent:])
		r, _ := gzip.NewReader(rdata)
		s, _ := ioutil.ReadAll(r)
		content = string(s)
	} else {
		content = string(buf[IndexOfContent:])
		compression = false
	}

	return FletaPacket{
		Command:     commend,
		Compression: compression,
		Content:     content,
	}
}

//ToJSON marshal json
func ToJSON(m interface{}) string {
	marshal, err := json.Marshal(m)
	if err != nil {
		log.Fatal("Cannot encode to JSON ", err)
	}
	// return base64.StdEncoding.EncodeToString(marshalstr)
	return string(marshal)
}

//FromJSON unmarshal json
func FromJSON(m interface{}, str string) interface{} {
	// by, err := base64.StdEncoding.DecodeString(str)
	// if err != nil {
	// 	log.Fatal("Cannot decode from JSON ", err)
	// }
	json.Unmarshal([]byte(str), m)
	return m
}

//ReadLoopFletaPacket TODO
func ReadLoopFletaPacket(pChan chan<- FletaPacket, conn net.Conn, exitGo <-chan bool) {
	var size = -1
	var buf, packetBuf []byte
	var packetEnd = false

	tmp := make([]byte, 256)
	for {
		n, err := conn.Read(tmp)
		if err != nil {
			if err != io.EOF {
				conn.Close()
				close(pChan)
			}
			break
		}

		if size == -1 {
			if len(buf) >= IndexOfContent && IsStartMagicWord(buf) {
				size = int(binary.LittleEndian.Uint32(buf[IndexOfSize : IndexOfSize+SizeOfSize]))
			} else if len(tmp) >= IndexOfContent && IsStartMagicWord(tmp) {
				size = int(binary.LittleEndian.Uint32(tmp[IndexOfSize : IndexOfSize+SizeOfSize]))
			} else {
				conn.Close()
				close(pChan)
			}
		}

		buf = append(buf, tmp[:n]...)

		var logBuf1 []byte
		var logBuf2 []byte
		if len(buf) == size+IndexOfContent {
			packetBuf = buf[:]
			buf = make([]byte, 0)
			packetEnd = true
		} else if len(buf) > size+IndexOfContent {
			logBuf1 = buf[:]
			packetBuf = buf[:size+IndexOfContent]
			buf = buf[size+IndexOfContent:]
			logBuf2 = buf[:]
			packetEnd = true
		}

		if packetEnd {
			tsize := int(binary.LittleEndian.Uint32(packetBuf[IndexOfSize:IndexOfSize+SizeOfSize])) + IndexOfContent
			if len(packetBuf) != tsize || !IsStartMagicWord(packetBuf) {
				fmt.Println(conn.LocalAddr().String(), " : ", size, ",", tsize, " packetBuf ", packetBuf, " : ", string(packetBuf))
				fmt.Println(conn.LocalAddr().String(), " :1 ", logBuf1, " : ", string(logBuf1))
				fmt.Println(conn.LocalAddr().String(), " :2 ", logBuf2, " : ", string(logBuf2))
				conn.Close()
				close(pChan)
				return
			}

			fp := FletaPacketToStruct(packetBuf)
			pChan <- fp

			select {
			case exit := <-exitGo:
				if exit {
					close(pChan)
					return
				}
			}

			packetBuf = make([]byte, 0)
			size = -1
			packetEnd = false
		}

	}

}

//ReadFletaPacket TODO
func ReadFletaPacket(conn net.Conn) (fpCh chan FletaPacket, exitCh chan bool) {
	pChan := make(chan FletaPacket, 1)
	exitChan := make(chan bool, 1)

	go ReadLoopFletaPacket(pChan, conn, exitChan)

	return pChan, exitChan

}
