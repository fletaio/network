package mockblock

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"fleta/flanetinterface"
	"fleta/mock"
	"fleta/mock/mocknet"
	util "fleta/samutil"
)

//error list
var (
	ErrNotFoundGenesisBlock = errors.New("Not found GenesisBlock")
	ErrInvalidHeight        = errors.New("InvalidHeight")
	ErrNotFoundBlock        = errors.New("Not Found Block")
	ErrNotFoundSeedNode     = errors.New("Not Found SeedNode")
)

//Block TODO
type Block struct {
	MakeBlockTime time.Time
	Height        int
	Addr          string
}

//Sync TODO
type Sync struct {
	sync.Mutex
	fi            FlanetImpl
	BlockList     []*Block
	tempBlockList []*Block
	blockAddrMap  map[string]*Block
}

//New is new
func New(fi FlanetImpl) *Sync {
	s := &Sync{
		BlockList:    make([]*Block, 0),
		blockAddrMap: make(map[string]*Block),
	}
	s.fi = fi

	// s.AddCommand("SYRSBLLS", func(conn net.Conn, fp util.FletaPacket) (exit bool, err error) {
	// 	var blocks []*Block
	// 	util.FromJSON(&blocks, fp.Content)
	// 	s.Log("%s", blocks)
	// 	if s.GetBlockHeight() == 0 {
	// 		if blocks[0].Height == 0 {
	// 			for _, b := range blocks {
	// 				s.PushBlock(b)
	// 			}
	// 		} else {
	// 			return false, ErrInvalidHeight
	// 		}
	// 	} else {
	// 		for _, b := range blocks {
	// 			s.PushBlock(b)
	// 		}

	// 	}
	// 	return false, nil
	// })
	// s.AddCommand("SYRQBLLT", func(conn net.Conn, fp util.FletaPacket) (exit bool, err error) {
	// 	bLen := len(s.BlockList)
	// 	requestLen, err := strconv.Atoi(fp.Content)
	// 	if err != nil {
	// 		requestLen = 0
	// 	}
	// 	if bLen > requestLen {
	// 		fletaPacket := util.FletaPacket{
	// 			Command: "SYRSBLLS",
	// 			Content: util.ToJSON(s.BlockList[requestLen:]),
	// 		}
	// 		p, err := fletaPacket.Packet()
	// 		if err != nil {
	// 			return false, err
	// 		}
	// 		conn.Write(p)
	// 	}
	// 	return false, nil
	// })

	return s
}

//GetConnList TODO
func (s *Sync) GetConnList() []net.Conn {
	conns := make([]net.Conn, 0)
	return conns
}

//VisualizationData TODO
func (s *Sync) VisualizationData() []string {
	blen := len(s.BlockList)
	if blen > 0 {
		lastBlock := s.BlockList[blen-1]
		return []string{fmt.Sprintf("%d %s", lastBlock.Height, lastBlock.Addr)}
	}
	return nil
}

//RegisteredRouter TODO
func (s *Sync) RegisteredRouter() error {
	return nil
}

//Close TODO
func (s *Sync) Close() {
}

//Start TODO
func (s *Sync) Start() {
	_, err := s.SeedNodeAddr()
	if err == ErrNotFoundSeedNode {
		s.requestBlockFromObserver()
	} else {
		go s.RequestBlock()
	}
}

//RequestBlock TODO
func (s *Sync) RequestBlock() {
	for {
		currentHeight := len(s.BlockList)
		fp := util.FletaPacket{
			Command: "SYRQBLLT",
			Content: strconv.Itoa(currentHeight),
		}
		if currentHeight == 0 {
			addr, err := s.SeedNodeAddr()
			if err != nil {
				// s.Error("%s", err)
				continue
			}
			conn, err := mocknet.Dial("tcp", addr, s.fi.Localhost())
			if err != nil {
				continue
			}
			pChan, _ := util.ReadFletaPacket(conn)
			go func() {
				fp := <-pChan

				if fp.Command == "SYRSBLLS" {
					var tBlockList []*Block
					util.FromJSON(&tBlockList, fp.Content)

					for _, block := range tBlockList {
						blen := len(s.BlockList)
						if blen == block.Height {
							s.PushBlock(block)
						}
					}
				}

				conn.Close()
			}()

			p, err := fp.Packet()
			if err != nil {
				// s.Error("%s", err)
				continue
			}
			conn.Write(p)

		} else {
			// s.ConsignmentCast(peerlist.PeerList{}.Location(), fp)
		}

		time.Sleep(time.Second * 30)
	}
}

func (s *Sync) requestBlockFromObserver() error {
	addr := s.fi.GetObserverNodeAddr()

	conn, err := mocknet.Dial("tcp", addr, s.fi.Localhost())
	if err != nil {
		return err
	}
	pChan, _ := util.ReadFletaPacket(conn)
	go func() {
		fp := <-pChan
		if fp.Command == "SYRSGEBL" {
			var block *Block
			util.FromJSON(&block, fp.Content)
			s.PushBlock(block)
		}
		conn.Close()
	}()

	fletaPacket := util.FletaPacket{
		Command: "OBRQGEBL",
	}

	p, err := fletaPacket.Packet()
	if err != nil {
		return err
	}
	conn.Write(p)

	return nil
}

func (s *Sync) putGenesisBlock(block Block) {
	s.blockAddrMap = make(map[string]*Block)
	s.BlockList = []*Block{&block}
	s.blockAddrMap[block.Addr] = &block
}

// Height TODO
func (s *Sync) Height() int {
	s.Lock()
	height := s.BlockList[len(s.BlockList)-1].Height
	s.Unlock()
	return height
}

//LastedBlock TODO
func (s *Sync) LastedBlock() *Block {
	index := len(s.BlockList) - 1
	return s.BlockList[index]
}

//PushBlock TODO
func (s *Sync) PushBlock(block *Block) bool {
	s.Lock()
	defer s.Unlock()
	if block.Height == 0 {
		s.pushBlock(block)
		return true
	}
	if bb := s.BlockList[block.Height-1]; bb != nil {
		if len(s.BlockList) == block.Height {
			s.pushBlock(block)
			return true
		}
	}
	return false
}

//MakeBlock TODO
func (s *Sync) MakeBlock(addr string) error {
	s.Lock()
	defer s.Unlock()
	length := len(s.BlockList)
	if length == 0 {
		return ErrNotFoundGenesisBlock
	}
	prevBlock := s.BlockList[length-1]
	if prevBlock.Height == length-1 {
		block := &Block{
			MakeBlockTime: time.Now(),
			Addr:          addr,
			Height:        length,
		}

		s.pushBlock(block)
		err := s.fi.SetBlockTime(addr, block.MakeBlockTime)
		if err != nil {
			return err
		}
		return nil
	}

	return ErrInvalidHeight
}

//CheckBlock TODO check other block height
func (s *Sync) pushBlock(block *Block) {
	if block.Height == 0 {
		s.blockAddrMap = make(map[string]*Block)
		s.BlockList = []*Block{block}
	} else {
		s.BlockList = append(s.BlockList, block)
	}
	s.blockAddrMap[block.Addr] = block

	if s.fi.GetNodeType() == flanetinterface.FormulatorNode {
		// err := s.fi.NewBlock(block)
		// if err != nil {
		// 	// s.Error("%s", err)
		// }
	}

}

//CheckBlock TODO check other block height
func (s *Sync) CheckBlock() {
	if len(s.BlockList) == 0 {
		// b.MakeGenesisBlock(b.Localhost())
	}
}

//FlanetImpl TODO
type FlanetImpl interface {
	GetObserverNodeAddr() string
	// NewBlock(*Block) error
	SeedNodeAddr() (string, error)
	GetNodeType() string
	SetBlockTime(string, time.Time) error
	Localhost() string
}

//GetMakeBlockTime is returns the time of the requested block
func (s *Sync) GetMakeBlockTime(addr string) (time.Time, error) {
	if block, ok := s.blockAddrMap[addr]; ok {
		return block.MakeBlockTime, nil
	}
	return time.Time{}, ErrNotFoundBlock
}

//SeedNodeAddr TODO
func (s *Sync) SeedNodeAddr() (string, error) {
	seedNodeAddr := util.Sha256HexInt(simulationdata.FormulatorNodeStartIndex)
	if s.fi.Localhost() == seedNodeAddr {
		return "", ErrNotFoundSeedNode
	}
	return seedNodeAddr, nil
}

//GetBlockHeight return height of block
func (s *Sync) GetBlockHeight() int {
	return len(s.BlockList)
}
