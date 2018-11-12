package mocknet

import (
	"sync"
	"time"
)

// NodeInfo has node infomation type, ID, data channel
type NodeInfo struct {
	Address       string
	ConnParamChan chan ConnParam
}

//Addr TODO
func (n *NodeInfo) Addr() string {
	return n.Address
}

//DetectedTime TODO
func (n *NodeInfo) DetectedTime() time.Time {
	return time.Now()
}

//BlockTime TODO
func (n *NodeInfo) BlockTime() time.Time {
	return time.Now()
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
