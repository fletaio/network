package hardstate

import (
	"errors"
	"sync"
	"time"

	"fleta/flanetinterface"
	"fleta/mock/mocknet"
	util "fleta/samutil"
)

//Errlist
var (
	ErrUseBeforeInit   = errors.New("Must call Init function use before ")
	ErrKeyIsNotValid   = errors.New("Map key type is Not Valid")
	ErrValueIsNotValid = errors.New("Map value type is Not Valid")
)

//HardStateImpl TODO
type HardStateImpl interface {
	RequestFormulatorList(string) error
	NewFormulator(node *flanetinterface.Node) error
	SendNewFormulators([]string) error
	PeerList() ([]flanetinterface.Node, error)
	Localhost() string
	Log(format string, msg ...interface{})
	Error(format string, msg ...interface{})
}

//HardState TODO
type HardState struct {
	HardStateImpl
	init             bool
	candidateNodeMap *sync.Map // key: string, value: flanetinterface.Node
	formulatorMap    *sync.Map // key: string, value: flanetinterface.Node
	sendedPeerMap    *sync.Map // key: string, value: bool
}

//Init TODO
func (h *HardState) Init(hi HardStateImpl) {
	h.HardStateImpl = hi
	h.candidateNodeMap = &sync.Map{}
	h.formulatorMap = &sync.Map{}
	h.sendedPeerMap = &sync.Map{}
	h.init = true
}

//IHardState TODO
type IHardState interface {
	// NewPeer(string)
	// FormulatorList() []flanetinterface.Node
}

//FormulatorAddrList TODO
func (h *HardState) FormulatorAddrList() ([]string, error) {
	if !h.init {
		return nil, ErrUseBeforeInit
	}
	nodes := make([]string, 0)
	h.formulatorMap.Range(func(Key, Value interface{}) bool {
		key, _ := h.keyValueObject(Key, Value)
		nodes = append(nodes, key)
		return true
	})

	return nodes, nil
}

//FormulatorList TODO
func (h *HardState) FormulatorList() ([]*flanetinterface.Node, error) {
	if !h.init {
		return nil, ErrUseBeforeInit
	}
	nodes := make([]*flanetinterface.Node, 0)
	h.formulatorMap.Range(func(Key, Value interface{}) bool {
		_, value := h.keyValueObject(Key, Value)
		nodes = append(nodes, value)
		return true
	})

	return nodes, nil
}

//NewPeer : send all list to new peer
func (h *HardState) NewPeer(pAddr string) error {
	if !h.init {
		return ErrUseBeforeInit
	}
	addrs := make([]string, 0)
	h.formulatorMap.Range(func(Key, Value interface{}) bool {
		key, _ := h.keyValueObject(Key, Value)
		addrs = append(addrs, key)
		return true
	})
	// h.Send(pAddr, addrs)
	return nil
}

func (h HardState) keyValueObject(_key, _value interface{}) (key string, value *flanetinterface.Node) {
	if key, ok := _key.(string); ok {
		if value, ok := _value.(*flanetinterface.Node); ok {
			return key, value
		}
	}
	return "", &flanetinterface.Node{}
}

//StartLoop TODO
func (h *HardState) StartLoop() {
	if !h.init {
		panic(ErrUseBeforeInit)
	}
	go func() {
		for {
			var hasCandidate bool
			h.candidateNodeMap.Range(func(Key, Value interface{}) bool {
				hasCandidate = true
				return false
			})
			formulatorNodes := make([]string, 0)
			h.formulatorMap.Range(func(Key, Value interface{}) bool {
				key, _ := h.keyValueObject(Key, Value)
				formulatorNodes = append(formulatorNodes, key)
				return true
			})

			h.Log("log : hasCandidate %t formulatorNodes %d\n", hasCandidate, len(formulatorNodes))
			if hasCandidate {
				h.checkCandidate()
			} else {
				list, err := h.PeerList()
				if err != nil {
					h.Error("%s", err)
				}
				for _, node := range list {
					addr := node.Addr()
					if _, ok := h.sendedPeerMap.Load(addr); !ok {
						h.sendedPeerMap.Store(addr, true)
						//TODO optimization
						h.RequestFormulatorList(addr)
						break
					}
				}
			}
			time.Sleep(time.Second * 10)
		}

	}()

}

func (h *HardState) checkCandidate() {
	nodes := make([]string, 0)
	comp := make([]string, 0)
	h.candidateNodeMap.Range(func(Key, Value interface{}) bool {
		key, node := h.keyValueObject(Key, Value)

		// conn, err := mocknet.Dial("tcp", node.Addr(), h.Localhost())
		// if err == nil {
		// 	conn.Close()
		// }

		conn, err := mocknet.Dial("tcp", node.Addr(), h.Localhost())
		if err != nil {
			h.Error("%s", err)
			return true
		}
		readyCh, pChan, _ := util.ReadFletaPacket(conn)
		go func() {
			timeout := make(chan bool)
			go func() {
				time.Sleep(time.Second * 5)
				timeout <- true
			}()

			select {
			case <-timeout:
				h.candidateNodeMap.Delete(Key)
				h.Log("timeout delete")
			case fp := <-pChan:

				conn.Close()

				if fp.Command == "FMHDRSFM" {
					h.candidateNodeMap.Delete(Key)
					if fp.Content == flanetinterface.FormulatorNode {
						if _, ok := h.formulatorMap.Load(Key); !ok {
							comp = append(comp, key)
							node.DetectedTime = time.Now()
							h.formulatorMap.Store(Key, node)
							nodes = append(nodes, node.Addr())
							err := h.NewFormulator(node)
							if err != nil {
								h.Error("%s", err)
							}
						}
					}

				}
			}

		}()

		fletaPacket := util.FletaPacket{
			Command: "FMHDRQFM",
		}

		p, err := fletaPacket.Packet()
		if err != nil {
			h.Error("%s", err)
			return true
		}
		<-readyCh
		conn.Write(p)

		return true
	})

	if len(nodes) > 0 {
		h.Log("send new formulator %d %s\n", len(nodes), nodes)
		h.SendNewFormulators(nodes)
	}

}

//GetNode TODO
func (h *HardState) GetNode(addr string) (node *flanetinterface.Node, okResult bool) {
	value, okResult := h.formulatorMap.Load(addr)
	var typeOk bool
	if node, typeOk := value.(*flanetinterface.Node); typeOk {
		return node, okResult
	}
	nilNode := &flanetinterface.Node{}
	if !typeOk {
		return nilNode, typeOk
	}
	return nilNode, okResult
}

//AddCandidateNodeAddr keep whole FormulatorNode
func (h *HardState) AddCandidateNodeAddr(nodes []string) error {
	if !h.init {
		return ErrUseBeforeInit
	}
	for _, addr := range nodes {
		_, ok := h.formulatorMap.Load(addr)
		if !ok {
			h.candidateNodeMap.Store(addr, &flanetinterface.Node{
				Address:  addr,
				NodeType: flanetinterface.FormulatorNode,
			})
		}
	}
	return nil
}

//AddCandidateNode keep whole FormulatorNode
func (h *HardState) AddCandidateNode(nodes []*flanetinterface.Node) error {
	if !h.init {
		return ErrUseBeforeInit
	}
	for _, node := range nodes {
		addr := node.Addr()
		_, ok := h.formulatorMap.Load(addr)
		if !ok {
			h.candidateNodeMap.Store(addr, node)
		}
	}
	return nil
}
