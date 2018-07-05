package hardstate

import (
	"errors"
	"fleta/flanetinterface"
	"fleta/mock/mocknet"
	"fleta/util"
	"sync"
	"time"
)

//Errlist
var (
	ErrKeyIsNotValid   = errors.New("Map key type is Not Valid")
	ErrValueIsNotValid = errors.New("Map value type is Not Valid")
)

//HardStateImpl TODO
type HardStateImpl interface {
	RequestFormulatorList(string) error
	Log(format string, msg ...interface{})
	PeerList() []flanetinterface.Node
	SelfNode() flanetinterface.Node
	Localhost() string

	Send(string, []string) error
}

//HardState TODO
type HardState struct {
	HardStateImpl
	candidateNodeMap *sync.Map // key: string, value: flanetinterface.Node
	formulatorMap    *sync.Map // key: string, value: flanetinterface.Node
	sendedPeerMap    *sync.Map // key: string, value: bool
}

//IHardState TODO
type IHardState interface {
	// NewPeer(string)
	// FormulatorList() []flanetinterface.Node
}

func (h *HardState) FormulatorAddrList() []string {
	nodes := make([]string, 0)
	h.formulatorMap.Range(func(Key, Value interface{}) bool {
		key, _ := h.keyValueObject(Key, Value)
		nodes = append(nodes, key)
		return true
	})

	return nodes
}

func (h *HardState) FormulatorList() []flanetinterface.Node {
	nodes := make([]flanetinterface.Node, 0)
	h.formulatorMap.Range(func(Key, Value interface{}) bool {
		_, value := h.keyValueObject(Key, Value)
		nodes = append(nodes, value)
		return true
	})

	return nodes
}

//NewPeer : send all list to new peer
func (h *HardState) NewPeer(pAddr string) error {
	addrs := make([]string, 0)
	h.formulatorMap.Range(func(Key, Value interface{}) bool {
		key, _ := h.keyValueObject(Key, Value)
		addrs = append(addrs, key)
		return true
	})
	h.Send(pAddr, addrs)
	return nil
}

func (h HardState) keyValueObject(_key, _value interface{}) (key string, value flanetinterface.Node) {
	if key, ok := _key.(string); ok {
		if value, ok := _value.(flanetinterface.Node); ok {
			return key, value
		}
	}
	return "", flanetinterface.Node{}
}

//Start TODO
func (h *HardState) Init(hs HardStateImpl) {
	h.HardStateImpl = hs
	h.candidateNodeMap = &sync.Map{}
	h.formulatorMap = &sync.Map{}
	h.sendedPeerMap = &sync.Map{}
}

//Start TODO
func (h *HardState) Start() {

	for {
		candidateNodes := make([]flanetinterface.Node, 0)
		h.candidateNodeMap.Range(func(Key, Value interface{}) bool {
			_, node := h.keyValueObject(Key, Value)
			candidateNodes = append(candidateNodes, node)
			return true
		})
		formulatorNodes := make([]flanetinterface.Node, 0)
		h.formulatorMap.Range(func(Key, Value interface{}) bool {
			_, node := h.keyValueObject(Key, Value)
			formulatorNodes = append(formulatorNodes, node)
			return true
		})

		h.Log("log : candidateNodes %d formulatorNodes %d", len(candidateNodes), len(formulatorNodes))

		if len(candidateNodes) == 0 {
			list := h.PeerList()
			for _, node := range list {
				addr := node.Addr()
				if _, ok := h.sendedPeerMap.Load(addr); !ok {
					h.sendedPeerMap.Store(addr, true)
					//TODO optimization
					h.RequestFormulatorList(addr)
					break
				}
			}

		} else {
			h.checkCandidate()
		}
		time.Sleep(time.Second * 10)
	}
}

func (h *HardState) checkCandidate() {
	h.candidateNodeMap.Range(func(Key, Value interface{}) bool {
		conn, err := mocknet.Dial("tcp", util.Sha256HexInt(0), h.Localhost())
		if err == nil {
			conn.Close()
			h.candidateNodeMap.Delete(Key)

			if node, ok := Value.(flanetinterface.Node); ok {
				node.DetectedTime = time.Now()
				h.formulatorMap.Store(Key, node)
			}
		}
		return true
	})
}

//AddFormulatorMap keep whole MasterNode
func (h *HardState) AddFormulatorMap(nodes []flanetinterface.Node) {
	for _, node := range nodes {
		h.formulatorMap.Store(node.Addr(), node)
	}
}

//AddCandidateNodeAddr keep whole MasterNode
func (h *HardState) AddCandidateNodeAddr(nodes []string) {
	for _, addr := range nodes {
		_, ok := h.formulatorMap.Load(addr)
		if !ok {
			h.candidateNodeMap.Store(addr, flanetinterface.Node{
				Address:  addr,
				NodeType: flanetinterface.MasterNode,
			})
		}
	}
}

//AddCandidateNode keep whole MasterNode
func (h *HardState) AddCandidateNode(nodes []flanetinterface.Node) {
	for _, node := range nodes {
		addr := node.Addr()
		_, ok := h.formulatorMap.Load(addr)
		if !ok {
			h.candidateNodeMap.Store(addr, node)
		}
	}
}
