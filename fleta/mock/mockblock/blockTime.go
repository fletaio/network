package mockblock

import (
	"time"
)

type BlockGen struct {
	GenTime       int64
	MakeBlockTime int64
	Addr          string
}

var BlockGenTime []*BlockGen

func Init() {
}

func Generation(node string) {
	BlockGenTime = append(BlockGenTime, &BlockGen{
		GenTime: time.Now().UnixNano(),
		Addr:    node,
	})
}

func MakeBlock(node string) {
	for _, b := range BlockGenTime {
		if b.Addr == node {
			b.MakeBlockTime = time.Now().UnixNano()
			return
		}
	}
}
