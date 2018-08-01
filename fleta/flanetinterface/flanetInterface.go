package flanetinterface

import "time"

//each node string
const (
	ObserverNode   = "OBSERVER_NODE"
	FormulatorNode = "FORMULATOR_NODE"
	NormalNode     = "NORMAL_NODE"
)

//port info
const (
	PeerPort         = 3000
	MinningGroupPort = 3001
)

//Node is sample node struct
type Node interface {
	Addr() string
	Type() string
	DetectedTime() time.Time
	BlockTime() time.Time
}
