package flanet

import (
	"bytes"
	"errors"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"fleta/flanet/hardstate"
	"fleta/flanet/network"
	"fleta/flanetinterface"
	"fleta/formulator"
	"fleta/message"
	"fleta/mininggroup"
	"fleta/mock"
	"fleta/mock/mockblock"
	"fleta/mock/mocknet"
	"fleta/packet"
	"fleta/peer"
	util "fleta/samutil"
)

//flanet error list
var (
	ErrUnregisteredCaster = errors.New("Unregistered Caster")
)

//Flanet is gateway of fleta
type Flanet struct {
	flanetID   int
	flanetType string

	hardstate hardstate.IHardState

	mg   *mininggroup.MiningGroup
	sync *mockblock.Sync
	pm   *peer.PeerManager

	msgChan chan message.Message
}

//NewFlanet is instance of flanet
func NewFlanet(i int, nodeType string) (*Flanet, error) {
	f := &Flanet{
		flanetID:   i,
		flanetType: nodeType,
	}

	h, err := hardstate.New(f, f.Localhost())
	if err != nil {
		return nil, err
	}
	f.hardstate = h

	return f, nil
}

func (f *Flanet) Mininggroup(mininggroup *mininggroup.MiningGroup) {
	f.mg = mininggroup
}
func (f *Flanet) Sync(sync *mockblock.Sync) {
	f.sync = sync
	f.touchSeedNode()
}
func (f *Flanet) PeerManager(pm *peer.PeerManager) {
	f.pm = pm
}

func (f *Flanet) VisualizationData() map[string][]string {
	data := make(map[string][]string)

	data["HSS"] = *f.hardstate.FormulatorList()
	data["HSS candidate"] = f.hardstate.CandidateList()
	if f.pm != nil {
		data["PEER"] = f.pm.PeerAddrList()
	}
	return data
}

func (f *Flanet) touchSeedNode() {
	addr, err := f.sync.SeedNodeAddr()
	if err != nil {
		f.Error("err : %s", err)
		return
	}
	f.DialTo(addr)
}

//Close Close obj
func (f *Flanet) flanetConsumer(msg message.Message) {
	mType, err := network.TypeOfMessage(msg)
	if err != nil {
		f.Error("err : %s", err)
		return
	}
	switch mType {
	case network.FormulatorListMessageType:
		if fl, ok := msg.(*network.FormulatorList); ok {
			if len(fl.List) > 1 {
				f.Debug("%d", len(fl.List))
			}
			for _, l := range fl.List {
				f.hardstate.NewNode(l)
			}

		}
	case network.AskFormulatorMessageType:
		localPeerAddr := f.Localhost() + ":" + strconv.Itoa(flanetinterface.PeerPort)
		af := network.NewAnswerFormulator(localPeerAddr, f.GetNodeType())

		payload := message.ToPayload(network.AnswerFormulatorMessageType, af)

		addr := msg.(*network.AskFormulator).Addr
		f.DialTo(addr)
		err := f.pm.Send(peer.NodeID(addr), &payload, packet.UNCOMPRESSED)
		if err != nil {
			f.pm.Delete(peer.NodeID(addr))
			if err == io.ErrClosedPipe {
			} else {
				f.Error("err : %s", err)
			}
		}

	case network.AnswerFormulatorMessageType:
		addr := msg.(*network.AnswerFormulator).Addr
		nodeType := msg.(*network.AnswerFormulator).NodeType
		f.hardstate.AnswerFormulator(addr, nodeType)
		f.pm.SetlimitPeerSize(len(*f.hardstate.FormulatorList()) / 3)
	}
}

//Close Close obj
func (f *Flanet) FlanetConsumer(msgChan chan message.Message) {
	f.msgChan = msgChan
	go func() {
		for {
			msg := <-msgChan
			go f.flanetConsumer(msg)
		}
	}()
}

//Close Close obj
func (f *Flanet) Close() {
}

//GetNodeType return node type
func (f *Flanet) GetNodeType() string {
	return f.flanetType
}

//GetFlanetID  return ID
func (f *Flanet) getFlanetID() int {
	return f.flanetID
}

//Localhost return localhost
func (f *Flanet) Localhost() string {
	return util.Sha256HexInt(f.flanetID)
}

//DialTo is wait of all accept
func (f *Flanet) DialTo(addr string) error {
	if addr == "" {
		return nil
	}
	if !strings.Contains(addr, ":") {
		addr += ":" + strconv.Itoa(flanetinterface.PeerPort)
	}
	nodeID := peer.NodeID(addr)
	_, err := f.pm.GetPeer(nodeID)
	if err != nil {
		if err == peer.ErrPeerNotExist {
			conn, err := mocknet.Dial("tcp", addr, f.Localhost()+":"+strconv.Itoa(flanetinterface.PeerPort))
			if err != nil {
				f.Error("err : %s", err)
				return err
			}
			addr = conn.RemoteAddr().String()
			err = f.pm.AddPeer(peer.NodeID(addr), conn)
			if err != nil {
				f.Error("err : %s", err)
			}
			f.hardstate.NewNode(addr)
			return nil
		}
		return err
	}
	return nil
}

//OpenFlanet is wait of all accept
func (f *Flanet) OpenFlanet() error {
	listen, err := mocknet.Listen("tcp", ":"+strconv.Itoa(flanetinterface.PeerPort))
	if err != nil {
		return err
	}

	for {
		conn, err := listen.Accept()
		if err != nil {
			continue
		}
		addr := conn.RemoteAddr().String()
		err = f.pm.AddPeer(peer.NodeID(addr), conn)
		if err != nil {
			f.Error("err : %s", err)
			conn.Close()
			continue
		}
		f.hardstate.NewNode(addr)
	}

}

//TODO MOCK DATA
func (f *Flanet) GetObserverNodeAddr() string {
	rand.Seed(time.Now().Unix())
	from := simulationdata.ObserverNodeStartIndex
	to := simulationdata.ObserverNodeStartIndex + simulationdata.ObserverNodeCount
	targetID := rand.Intn(to-from) + from
	return util.Sha256HexInt(targetID)
}

//SaveNewNode TODO
func (f *Flanet) SaveNewNode(addr string) ([]byte, error) {
	node := &formulator.Node{
		Address: addr,
		// Detected: time.Now(),
	}
	if f.mg != nil {
		f.mg.NewFormulator(node)
	}
	var buffer bytes.Buffer
	if _, err := node.WriteTo(&buffer); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

//SpreadNewNodes TODO
func (f *Flanet) SpreadNewNodes(addr string) ([]string, error) {
	responseList := network.NewFormulatorList([]string{addr})
	payload := message.ToPayload(network.FormulatorListMessageType, responseList)

	nodeIDs := f.pm.Broadcast(&payload)

	spreadedList := make([]string, len(nodeIDs))
	for i, spreaded := range nodeIDs {
		spreadedList[i] = string(spreaded)
	}

	return spreadedList, nil
}

//AskFormulator check the target node is formulator
func (f *Flanet) AskFormulator(addr string) (isLocalhost bool) {
	localhost := f.Localhost() + ":" + strconv.Itoa(flanetinterface.PeerPort)
	if addr == localhost {
		return true
	}
	askf := network.NewAskFormulator(localhost)
	payload := message.ToPayload(network.AskFormulatorMessageType, askf)

	f.DialTo(addr)
	f.pm.Send(peer.NodeID(addr), &payload, packet.UNCOMPRESSED)
	return false
}

//SendToNodeList TODO
func (f *Flanet) SendToNodeList(addr string, list []string) error {

	responseList := network.NewFormulatorList(list)
	payload := message.ToPayload(network.FormulatorListMessageType, responseList)

	f.DialTo(addr)
	f.pm.Send(peer.NodeID(addr), &payload, packet.UNCOMPRESSED)

	return nil
}

/*
Formulator Impls
*/
func (f *Flanet) SetBlockTime(addr string, t time.Time) error {
	return f.hardstate.UpdateNode(addr, func(v []byte) ([]byte, error) {
		var node *formulator.Node

		r := bytes.NewReader(v)
		_, err := node.ReadFrom(r)
		if err != nil {
			return nil, err
		}
		// node.Block = t
		return v, nil
	})
}

/*
Block Impls
*/

//IBlock is requested to implement a list of Block functions
type ISync interface {
	MakeBlock(addr string) error
	GetMakeBlockTime(addr string) (time.Time, error)
	SeedNodeAddr() (string, error)
	GetBlockHeight() int
}

/*
Block Impls
*/
func (f *Flanet) MakeBlock() error {
	if f.sync != nil {
		return f.sync.MakeBlock(f.Localhost())
	}
	return nil
}

func (f *Flanet) GetMakeBlockTime(addr string) (time.Time, error) {
	if f.sync != nil {
		return f.sync.GetMakeBlockTime(addr)
	}
	return time.Time{}, nil
}

func (f *Flanet) SeedNodeAddr() (string, error) {
	if f.sync != nil {
		return f.sync.SeedNodeAddr()
	}
	return "", nil
}
func (f *Flanet) GetBlockHeight() int {
	if f.sync != nil {
		return f.sync.GetBlockHeight()
	}
	return -1
}
