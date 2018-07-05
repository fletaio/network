package network

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"fleta/flanetinterface"
	"fleta/mock/mockblock"
	"fleta/util"
)

//IFleta Fleta interface
type IFleta interface {
	Start(i int, nodeType string)
	NewFleta() IFleta
}

var fletaTest IFleta

//Fleta is set Fleta
func Fleta(_fleta IFleta) {
	fletaTest = _fleta
}

const (
	guardNodeCount  = 0
	seedNodeCount   = 1
	masterNodeCount = 50
	normalNodeCount = 0
)

const (
	guardNodeStartIndex  = 0
	seedNodeStartIndex   = guardNodeCount
	masterNodeStartIndex = seedNodeStartIndex + seedNodeCount
	normalNodeStartIndex = masterNodeStartIndex + masterNodeCount
)

const (
	masterNodePeerSize = 3
	peerSize           = 54
)

// NodeInfo has node infomation type, ID, data channel
type NodeInfo struct {
	NodeType      string
	NodeID        string
	ConnParamChan chan ConnParam
	ft            IFleta
}

var nodeMap *sync.Map

func init() {
	nodeMap = &sync.Map{}
}

//LoadNodeMap is safe-thread map Load()
func LoadNodeMap(key string) NodeInfo {
	if value, ok := nodeMap.Load(key); ok {
		if val, ok := value.(NodeInfo); ok {
			return val
		}
	}
	return NodeInfo{}
}

//StoreNodeMap is safe-thread map Store()
func StoreNodeMap(key string, n NodeInfo) {
	nodeMap.Store(key, n)
}

//Run is start mocknetwork
func Run() {
	totalCount := guardNodeCount + seedNodeCount + masterNodeCount + normalNodeCount
	i := 0
	for i := 0; i < totalCount; i++ {
		appendNode(i)
	}
	for i = 0; i < totalCount; i++ {
		go func(i int) {
			LoadNodeMap(util.Sha256HexString(strconv.Itoa(i))).ft.Start(i, itoNodeType(i))
		}(i)
	}

	// for j := seedNodeStartIndex; j < masterNodeStartIndex; j++ {
	// 	go pushPeerList(j)
	// }
}

func itoNodeType(i int) string {
	if i < seedNodeStartIndex {
		return flanetinterface.GuardNode
	} else if i < masterNodeStartIndex {
		return flanetinterface.SeedNode
	} else if i < normalNodeStartIndex {
		return flanetinterface.MasterNode
	} else {
		return flanetinterface.NormalNode
	}
}

func pushPeerList(localhost string) error {
	cRead, cWrite := RegistDial("tcp", localhost, localhost)

	nodes := make([]flanetinterface.Node, 0)
	totalCount := guardNodeCount
	i := 0
	for ; i < totalCount; i++ {
		node := flanetinterface.Node{
			Address:  util.Sha256HexString(strconv.Itoa(i)),
			NodeType: flanetinterface.GuardNode,
		}
		nodes = append(nodes, node)
	}
	totalCount += seedNodeCount
	for ; i < totalCount; i++ {
		node := flanetinterface.Node{
			Address:  util.Sha256HexString(strconv.Itoa(i)),
			NodeType: flanetinterface.SeedNode,
		}
		nodes = append(nodes, node)
	}
	totalCount += masterNodeCount
	for ; i < totalCount; i++ {
		node := flanetinterface.Node{
			Address:  util.Sha256HexString(strconv.Itoa(i)),
			NodeType: flanetinterface.MasterNode,
		}
		nodes = append(nodes, node)
	}
	totalCount += normalNodeCount
	for ; i < totalCount; i++ {
		node := flanetinterface.Node{
			Address:  util.Sha256HexString(strconv.Itoa(i)),
			NodeType: flanetinterface.NormalNode,
		}
		nodes = append(nodes, node)
	}

	seri := util.ToJSON(nodes)

	fp := util.FletaPacket{
		Command:     "PLRTPELT",
		Compression: false,
		Content:     seri,
	}

	packet, err := fp.Packet()
	if err != nil {
		return err
	}

	cWrite.Write(packet)

	if err := cWrite.Close(); err != nil {
		return err
	}
	if err := cRead.Close(); err != nil {
		return err
	}

	return nil
}

func appendNode(i int) {
	appendNodeAddress(itoNodeType(i), util.Sha256HexString(strconv.Itoa(i)))
}
func appendNodeAddress(nodeName string, addr string) {
	nodeInfo := NodeInfo{
		NodeType: nodeName,
		NodeID:   addr,
		ft:       fletaTest.NewFleta(),
	}

	if nodeName == flanetinterface.MasterNode {
		mockblock.Generation(addr)
	}

	StoreNodeMap(nodeInfo.NodeID, nodeInfo)

}

// SeedNode test
func SeedNode() NodeInfo {
	return LoadNodeMap(util.Sha256HexInt(5))
}

func childIDGen(nodeID string, index int) string {
	runes := []rune(nodeID)

	var strBuilder strings.Builder
	strBuilder.WriteString(string(runes[index:]))
	strBuilder.WriteString(string(runes[:index]))

	return util.Sha256HexString(strBuilder.String())
}

//ConnParam has Reader, Writer, network, address
type ConnParam struct {
	Reader      *io.PipeReader
	Writer      *io.PipeWriter
	NetworkType string
	Address     string
	DialHost    string
}

//RegistDial is temp store reader and writer
func RegistDial(networkType, address string, localhost string) (_cRead *io.PipeReader, _cWrite *io.PipeWriter) {
	for LoadNodeMap(address).ConnParamChan == nil {
		time.Sleep(100 * time.Millisecond)
	}
	sRead, cWrite := io.Pipe()
	cRead, sWrite := io.Pipe()

	connParam := ConnParam{
		Reader:      sRead,
		Writer:      sWrite,
		NetworkType: networkType,
		Address:     address,
		DialHost:    localhost,
	}
	LoadNodeMap(address).ConnParamChan <- connParam

	return cRead, cWrite
}

//RegistAccept is temp store reader and writer
func RegistAccept(addr string) (node NodeInfo) {
	if LoadNodeMap(addr).NodeID == "" {
		appendNodeAddress(addr, addr)
	}

	node = LoadNodeMap(addr)
	node.ConnParamChan = make(chan ConnParam, 256)
	StoreNodeMap(addr, node)

	return node
}

//Ping is ping to input address
func Ping(address string) {
	my := GetMainID()
	log.Println(my)
	log.Println(address)

	myDecoded, err := hex.DecodeString(my)
	if err != nil {
		log.Fatal(err)
	}
	addressDecoded, err := hex.DecodeString(address)
	if err != nil {
		log.Fatal(err)
	}

	data := binary.BigEndian.Uint32(myDecoded[:4]) - binary.BigEndian.Uint32(addressDecoded[:4])

	fmt.Printf("%s\n", myDecoded)
	fmt.Printf("%s\n", addressDecoded)
	fmt.Printf("%d\n", data)
}

//GetMainID is tracking call-stack until find "fleta.(*Fleta).Start" and return that ID
//HARD DEFENDENCE ON MOCKNETWORK
func GetMainID() string {
	buf := make([]byte, 1<<16)
	runtime.Stack(buf, true)
	strs := strings.Split(string(buf), "\n")
	re := regexp.MustCompile("fleta.\\(\\*Fleta\\).Start[^(]*\\(")
	for _, str := range strs {
		if strings.Contains(str, "fleta.(*Fleta).Start") {
			str = re.ReplaceAllLiteralString(str, "")
			str = strings.TrimRight(str, ")")
			ids := strings.Split(str, ",")
			if len(ids) >= 2 {
				var num int64
				num, err := strconv.ParseInt(strings.TrimPrefix(ids[1], " 0x"), 16, 32)
				if err != nil {
					log.Fatal(err)
				}
				i := int(num)
				return util.Sha256HexInt(i)

			}
		}
	}

	return util.Sha256HexInt(-1)
}
