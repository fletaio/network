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
	util "fleta/samutil"
	"fleta/samutil/concentrator"
)

//PeerList is sample peerlist struct
type PeerList struct {
	peerMap *sync.Map
	connMap *sync.Map
	fi      FlanetImpl
	concentrator.Caster
}

var peerListSize = 15

//error message list
var (
	ErrPeerNotExist = errors.New("ErrPeerNotExist")
	ErrConnNotExist = errors.New("ErrConnNotExist")
	ErrConnNotValid = errors.New("ErrConnNotValid")
)

//Location character of NetCaster
func (pl PeerList) Location() string {
	return "PL"
}

//New TODO
func New(fi FlanetImpl) *PeerList {
	pl := &PeerList{}
	pl.peerMap = &sync.Map{}
	pl.connMap = &sync.Map{}
	pl.fi = fi

	pl.Caster.Init(pl)

	pl.addProcessCommand()

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
		if plen < peerListSize {
			err := pl.requestPeerList()
			if err != nil {
				pl.Error("%s", err)
			}
			if plen == 0 {
				time.Sleep(time.Second * 1)
			} else {
				time.Sleep(time.Second * 5)
			}
		} else {
			pl.Log("plen : %d", plen)
			break
		}
	}
}

func (pl *PeerList) requestPeerList() error {
	seedAddr, err := pl.fi.SeedNodeAddr()
	if err != nil {
		return err
	}
	conn, err := mocknet.Dial("tcp", seedAddr, pl.Localhost())
	if err != nil {
		return err
	}
	readyCh, pChan, _ := util.ReadFletaPacket(conn)
	go func() {
		fp := <-pChan
		_, err := pl.RunCommand(conn, fp)
		conn.Close()
		if err != nil {
			pl.Error("%s fp : %s", err, fp)
		}
	}()

	fletaPacket := util.FletaPacket{
		Command: "PLRTRQPL",
		Content: pl.GetHint(),
	}

	p, err := fletaPacket.Packet()
	if err != nil {
		return err
	}
	<-readyCh
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
	readyCh, pChan, exitChan := util.ReadFletaPacket(conn)
	<-readyCh
	for {
		fp := <-pChan
		exit, err := pl.RunCommand(conn, fp)
		if err != nil {
			pl.Error("%s", err)
		}
		exitChan <- exit
	}

}

//PushPeerList is push peerlist
func (pl *PeerList) PushPeerList(peerlist []flanetinterface.Node) {
	for _, node := range peerlist {
		if node.NodeType == flanetinterface.FormulatorNode {
			err := pl.fi.DetectFormulatorNode(node)
			if err != nil {
				pl.Error("%s", err)
			}
		}

		if _, ok := pl.peerMap.Load(node.Address); !ok {
			pl.peerMap.Store(node.Address, node)
			if mapLen(pl.peerMap) < peerListSize {
				go pl.makeConn(node.Address)
			}
		}
	}
}

//BroadCastToFormulator is broadcast to formulator
func (pl *PeerList) BroadCastToFormulator(fp util.FletaPacket) error {
	conTarget := map[string]bool{}
	p, err := fp.Packet()
	if err != nil {
		return err
	}
	pl.connMap.Range(func(key, value interface{}) bool {
		if conn, ok := value.(net.Conn); ok {
			if nodeInter, ok := pl.peerMap.Load(key); ok {
				if node, ok := nodeInter.(flanetinterface.Node); ok {
					if node.Type() == flanetinterface.FormulatorNode {
						key := fmt.Sprintf("%s%s", conn.LocalAddr().String(), conn.RemoteAddr().String())
						if _, ok := conTarget[key]; !ok {
							conTarget[key] = true
							conn.Write(p)
						}
					}
				}
			}
		}
		return true
	})
	return nil
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
func (pl *PeerList) VisualizationData() []string {

	values := make([]string, 0)
	pl.peerMap.Range(func(key, value interface{}) bool {
		if node, ok := value.(flanetinterface.Node); ok {
			values = append(values, node.Addr())
		}
		return true
	})
	return values

}

func (pl *PeerList) addProcessCommand() {
	pl.AddCommand("PLMAKECN", func(conn net.Conn, fp util.FletaPacket) (exit bool, err error) {
		//shack hand
		return false, nil
	})
	pl.AddCommand("PLRTPELT", func(conn net.Conn, fp util.FletaPacket) (exit bool, err error) {
		nodes := make([]flanetinterface.Node, 0)
		util.FromJSON(&nodes, fp.Content)
		pl.PushPeerList(nodes)
		return false, nil
	})
	pl.AddCommand("PLRTRQPL", func(conn net.Conn, fp util.FletaPacket) (exit bool, err error) {
		addr := conn.RemoteAddr().String()
		nodeType := fp.Content

		_, ok := pl.peerMap.Load(pl.Localhost())
		if !ok {
			pl.PushPeerList([]flanetinterface.Node{flanetinterface.Node{
				Address:  pl.Localhost(),
				NodeType: pl.GetHint(),
			}})
		}
		_, ok = pl.peerMap.Load(addr)
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

			return false, nil
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

		sendfp := util.FletaPacket{
			Command:     "PLRTPELT",
			Compression: false,
			Content:     util.ToJSON(arrNode),
		}
		p, _ := sendfp.Packet()
		conn.Write(p)
		return false, nil
	})
}

//FlanetImpl TODO
type FlanetImpl interface {
	NewPeer(address string) error
	DetectFormulatorNode(flanetinterface.Node) error
	SeedNodeAddr() (string, error)
}

// //IPeerList TODO
// type IPeerList interface {
// 	PeerList() []flanetinterface.Node
// 	Send(address string, fp util.FletaPacket) error
// 	BroadCastToFormulator(fp util.FletaPacket) error
// }

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
