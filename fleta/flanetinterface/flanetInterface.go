package flanetinterface

import (
	"time"
)

//each node string
const (
	ObserverNode   = "OBSERVER_NODE"
	FormulatorNode = "FORMULATOR_NODE"
	NormalNode     = "NORMAL_NODE"
)

//Node is sample node struct
type Node struct {
	ID           int
	Address      string
	NodeType     string
	DetectedTime time.Time
	BlockTime    time.Time
}

func (n *Node) Addr() string {
	return n.Address
}
func (n *Node) Type() string {
	return n.NodeType
}
