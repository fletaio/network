package util

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
const MagicWord = "FLET"

//MagicWordByte is FletaPacket's MagicWordByte
var MagicWordByte = []byte{70, 76, 69, 84}
var (
	//ErrInvalidPacket TODO
	ErrInvalidPacket = errors.New("Invalid packet")
)

//IsStartMagicWord is check the packet start with magic word
func IsStartMagicWord(tmp []byte) bool {
	if tmp[0] == MagicWordByte[0] && tmp[1] == MagicWordByte[1] && tmp[2] == MagicWordByte[2] && tmp[3] == MagicWordByte[3] {
		return true
	}
	return false
}

//FletaPacket struct
//MagicWord(32), Command(64), Compression(8), Size(32), Content
type FletaPacket struct {
	Command     string
	Compression bool
	Content     string
}

// ErrPacket is returned for wrong packet
var ErrPacket error = &PacketError{}

// PacketError is returned for wrong packet
type PacketError struct {
	magicWord bool
	command   bool
}

// Implement the net.Error interface.
func (e *PacketError) Error() string   { return "i/o timeout" }
func (e *PacketError) MagicWord() bool { return e.magicWord }
func (e *PacketError) Command() bool   { return e.command }

func (fp *FletaPacket) validate() error {
	if len(MagicWord) != 4 {
		return &PacketError{
			magicWord: true,
			command:   false,
		}
	}
	if len(fp.Command) != 8 {
		return &PacketError{
			magicWord: false,
			command:   true,
		}
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
	sizeByte := make([]byte, 4)
	binary.LittleEndian.PutUint32(sizeByte, size)
	packet = append(packet, sizeByte...)
	packet = append(packet, content...)

	return packet, nil
}

//FletaPacketToStruct is marshal packet
func FletaPacketToStruct(buf []byte) FletaPacket {
	commend := string(buf[4:12])
	var compression bool
	var content string
	if buf[12] == 1 {
		compression = true
		// data, _ := base64.StdEncoding.DecodeString(string(buf[17:]))
		rdata := bytes.NewReader(buf[17:])
		r, _ := gzip.NewReader(rdata)
		s, _ := ioutil.ReadAll(r)
		content = string(s)
	} else {
		content = string(buf[17:])
		compression = false
	}

	return FletaPacket{
		Command:     commend,
		Compression: compression,
		Content:     content,
	}
}

// go json encoder
func ToJSON(m interface{}) string {
	marshal, err := json.Marshal(m)
	if err != nil {
		log.Fatal("Cannot encode to JSON ", err)
	}
	return string(marshal)
}

// go json decoder
func FromJSON(m interface{}, str string) interface{} {
	// by, err := base64.StdEncoding.DecodeString(str)
	// if err != nil {
	// 	log.Fatal("Cannot decode from JSON ", err)
	// }
	json.Unmarshal([]byte(str), m)
	return m
}

func ReadLoopFletaPacket(pChan chan<- FletaPacket, conn net.Conn, readyToReadChan chan<- bool) {
	var size = -1
	var buf, packetBuf []byte
	var packetEnd = false

	tmp := make([]byte, 256)
	readyToReadChan <- true
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
			if len(buf) >= 17 && IsStartMagicWord(buf) {
				size = int(binary.LittleEndian.Uint32(buf[13:17]))
			} else if len(tmp) >= 17 && IsStartMagicWord(tmp) {
				size = int(binary.LittleEndian.Uint32(tmp[13:17]))
			} else {
				conn.Close()
				close(pChan)
			}
		}

		buf = append(buf, tmp[:n]...)

		var logBuf1 []byte
		var logBuf2 []byte
		if len(buf) == size+17 {
			packetBuf = buf[:]
			buf = make([]byte, 0)
			packetEnd = true
		} else if len(buf) > size+17 {
			logBuf1 = buf[:]
			packetBuf = buf[:size+17]
			buf = buf[size+17:]
			logBuf2 = buf[:]
			packetEnd = true
		}

		if packetEnd {
			tsize := int(binary.LittleEndian.Uint32(packetBuf[13:17])) + 17
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
			if fp.Command == "MGEXLOOP" {
				return
			}

			packetBuf = make([]byte, 0)
			size = -1
			packetEnd = false
		}

	}
}
