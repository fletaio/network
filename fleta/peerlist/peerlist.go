package peerlist

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

	"fleta/flanetinterface"
	"fleta/mock/mocknet"
	"fleta/util"
	"fleta/util/netcaster"
)

//PeerList is sample peerlist struct
type PeerList struct {
	peerMap *sync.Map
	connMap *sync.Map
	pi      PeerListImpl
	netcaster.NetCaster
}

var peerListSize = 25
var (
	ErrPeerNotExist = errors.New("ErrPeerNotExist")
	ErrConnNotExist = errors.New("ErrConnNotExist")
	ErrConnNotValid = errors.New("ErrConnNotValid")
)

func Location() string {
	return "PL"
}
func (pl *PeerList) Location() string {
	return Location()
}

//New TODO
func New(pi PeerListImpl, cr *netcaster.CastRouter, hint string) *PeerList {
	pl := &PeerList{}
	pl.peerMap = &sync.Map{}
	pl.connMap = &sync.Map{}
	pl.pi = pi

	pl.NetCaster = netcaster.NetCaster{
		pl, cr, hint,
	}

	return pl
}

func mapLen(m *sync.Map) int {
	count := 0
	m.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

//RenewPeerList is keep peer list
func (pl *PeerList) RenewPeerList() {
	for {
		plen := mapLen(pl.peerMap)
		if plen == 0 {
			pl.requestPeerList()
			time.Sleep(time.Second * 1)
		} else if plen < peerListSize {
			pl.requestPeerList()
			time.Sleep(time.Second * 5)
		} else {
			pl.Log("plen : %d", plen)
			break
		}
	}
}

func (pl *PeerList) requestPeerList() error {
	conn, err := mocknet.Dial("tcp", util.Sha256HexInt(0), pl.Localhost())
	if err != nil {
		return err
	}
	readyToReadChan := make(chan bool, 256)
	go func() {
		fChan := make(chan util.FletaPacket, 1)
		go util.ReadLoopFletaPacket(fChan, conn, readyToReadChan)
		fp := <-fChan
		if fp.Command != "" {
			pl.ProcessPacket(conn, fp)
		}
		conn.Close()
	}()

	fletaPacket := util.FletaPacket{
		Command: "PLRTRQPL",
		Content: pl.Hint,
	}

	p, err := fletaPacket.Packet()
	if err != nil {
		return err
	}
	<-readyToReadChan
	conn.Write(p)
	return nil
}

//makeConn is push peerlist
func (pl *PeerList) makeConn(addr string) error {
	conn, err := mocknet.Dial("tcp", addr, pl.Localhost())
	if err != nil {
		return err
	}
	pl.addConn(conn)

	fp := util.FletaPacket{
		Command: "PLMAKECN",
	}
	p, err := fp.Packet()
	if err != nil {
		return err
	}
	conn.Write(p)

	return nil

}

func (pl *PeerList) addConn(conn net.Conn) {
	key := conn.RemoteAddr().String()
	if _, ok := pl.peerMap.Load(key); ok {
		if _, ok := pl.connMap.Load(key); !ok {
			pl.connMap.Store(key, conn)
			go pl.readPacket(conn)
		}
	}
}

//readPacket is push peerlist
func (pl *PeerList) readPacket(conn net.Conn) {
	readyToReadChan := make(chan bool)
	fChan := make(chan util.FletaPacket, 1)
	go util.ReadLoopFletaPacket(fChan, conn, readyToReadChan)
	<-readyToReadChan
	for {
		fp := <-fChan
		if fp.Command == "" {
			break
		}
		pl.ProcessPacket(conn, fp)

	}

}

//PushPeerList is push peerlist
func (pl *PeerList) PushPeerList(peerlist []flanetinterface.Node) {
	for _, node := range peerlist {
		if node.NodeType == flanetinterface.MasterNode {
			pl.pi.DetectMasterNode(node)
		}

		if _, ok := pl.peerMap.Load(node.Address); !ok {
			pl.peerMap.Store(node.Address, node)
			if pl.Hint != flanetinterface.SeedNode {
				go pl.makeConn(node.Address)
				if mapLen(pl.peerMap) >= peerListSize {
					break
				}
			}
		}
	}
}

//GetConnList is get conn obj
func (pl *PeerList) GetConnList() []net.Conn {
	conns := make([]net.Conn, 0)
	conTarget := map[string]bool{}
	pl.connMap.Range(func(key, value interface{}) bool {
		if conn, ok := value.(net.Conn); ok {
			key := fmt.Sprintf("%s%s", conn.LocalAddr().String(), conn.RemoteAddr().String())
			if _, ok := conTarget[key]; !ok {
				conTarget[key] = true
				conns = append(conns, conn)
			}
		}
		return true
	})

	return conns
}

//ProcessPacket is handling process packet
func (pl *PeerList) ProcessPacket(conn net.Conn, p util.FletaPacket) error {
	switch p.Command {
	case "PLMAKECN":
		//shack hand
	case "PLRTPELT":
		nodes := make([]flanetinterface.Node, 0)
		util.FromJSON(&nodes, p.Content)
		pl.PushPeerList(nodes)
	case "PLRTRQPL":
		addr := conn.RemoteAddr().String()
		nodeType := p.Content

		_, ok := pl.peerMap.Load(addr)
		if !ok {
			pl.PushPeerList([]flanetinterface.Node{flanetinterface.Node{
				Address:  addr,
				NodeType: nodeType,
			}})
		}

		pplLen := mapLen(pl.peerMap)
		if pplLen <= 1 {
			fp := util.FletaPacket{
				Command:     "PLRTPELT",
				Compression: false,
				Content:     util.ToJSON(make([]flanetinterface.Node, 0)),
			}
			p, _ := fp.Packet()
			conn.Write(p)

			return nil
		}
		nodeLen := peerListSize
		if pplLen <= peerListSize {
			nodeLen = pplLen - 1
		}

		keys := make([]string, 0)
		i := 0
		pl.peerMap.Range(func(key, value interface{}) bool {
			if k, ok := key.(string); ok {
				if addr != k {
					keys = append(keys, k)
					i++
				}
			}
			return true
		})

		rand.Seed(time.Now().UTC().UnixNano())
		var nodes = map[string]flanetinterface.Node{}
		for targetI := rand.Intn(len(keys)); len(nodes) < nodeLen; targetI = rand.Intn(len(keys)) {
			val, _ := pl.peerMap.Load(keys[targetI])
			if node, ok := val.(flanetinterface.Node); ok {
				nodes[node.Address] = node
			}
		}

		arrNode := []flanetinterface.Node{}
		for _, node := range nodes {
			arrNode = append(arrNode, node)
		}

		fp := util.FletaPacket{
			Command:     "PLRTPELT",
			Compression: false,
			Content:     util.ToJSON(arrNode),
		}
		p, _ := fp.Packet()
		conn.Write(p)
	default:
		cs := pl.LocalRouter(p.Command[:2])
		if cs != nil {
			return cs.ProcessPacket(conn, p)
		}
	}
	return nil
}

//PeerListImpl TODO
type PeerListImpl interface {
	NewPeer(address string)
	DetectMasterNode(flanetinterface.Node)
}

//IPeerList TODO
type IPeerList interface {
	PeerList() []flanetinterface.Node
	Send(address string, fp util.FletaPacket) error
}

//PeerList TODO
func (pl *PeerList) PeerList() []flanetinterface.Node {
	values := make([]flanetinterface.Node, 0)
	pl.peerMap.Range(func(key, value interface{}) bool {
		if node, ok := value.(flanetinterface.Node); ok {
			values = append(values, node)
		}
		return true
	})
	return values
}

//PeerSend TODO
func (pl *PeerList) PeerSend(address string, fp util.FletaPacket) error {
	if _, ok := pl.peerMap.Load(address); ok {
		if val, ok := pl.connMap.Load(address); ok {
			if conn, ok := val.(net.Conn); ok {
				p, err := fp.Packet()
				if err != nil {
					return err
				}
				conn.Write(p)
				return nil
			}
			return ErrConnNotValid
		}
		return ErrConnNotExist
	}
	return ErrPeerNotExist
}
