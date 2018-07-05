package flanetinterface

import (
	"time"
)

const (
	//GuardNode string
	GuardNode = "GUARD_NODE"
	//SeedNode string
	SeedNode = "SEED_NODE"
	//MasterNode string
	MasterNode = "MASTER_NODE"
	//NormalNode string
	NormalNode = "NORMAL_NODE"
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
