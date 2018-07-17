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
	"fleta/peerlist"
	util "fleta/samutil"
	"fleta/samutil/concentrator"
)

//error list
var (
	ErrNotFoundGenesisBlock = errors.New("Not found GenesisBlock")
	ErrInvalidHeight        = errors.New("InvalidHeight")
	ErrNotFoundBlock        = errors.New("Not Found Block")
	ErrNotFoundSeedNode     = errors.New("Not Found SeedNode")
)

type Block struct {
	MakeBlockTime time.Time
	Height        int
	Addr          string
}

type Sync struct {
	sync.RWMutex
	fi           FlanetImpl
	BlockList    []*Block
	blockAddrMap map[string]*Block
	concentrator.Caster
}

func New(fi FlanetImpl) *Sync {
	s := &Sync{
		BlockList:    make([]*Block, 0),
		blockAddrMap: make(map[string]*Block),
	}
	s.fi = fi

	s.Caster.Init(s)

	s.AddCommand("SYRQBLLT", func(conn net.Conn, fp util.FletaPacket) (exit bool, err error) {
		bLen := len(s.BlockList)
		requestLen, err := strconv.Atoi(fp.Content)
		if err != nil {
			requestLen = 0
		}
		if bLen > requestLen {
			fletaPacket := util.FletaPacket{
				Command: "SYRSBLLS",
				Content: util.ToJSON(s.BlockList[requestLen:]),
			}
			p, err := fletaPacket.Packet()
			if err != nil {
				return false, err
			}
			conn.Write(p)
		}
		return false, nil
	})

	return s
}

//Location TODO
func (s Sync) Location() string {
	return "SY"
}

//GetConnList TODO
func (s *Sync) GetConnList() []net.Conn {
	conns := make([]net.Conn, 0)
	return conns
}
func (s *Sync) VisualizationData() []string {
	blen := len(s.BlockList)
	if blen > 0 {
		lastBlock := s.BlockList[blen-1]
		return []string{fmt.Sprintf("%d %s", lastBlock.Height, lastBlock.Addr)}
	}
	return nil
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
		list, err := s.fi.PeerList()
		if err != nil {
			s.Error("%s", err)
		}
		if len(list) > 0 {
			break
		}
		time.Sleep(time.Second)
	}

	for {
		fp := util.FletaPacket{
			Command: "SYRQBLLT",
		}
		if len(s.BlockList) == 0 {
			addr, err := s.fi.SeedNodeAddr()
			if err != nil {
				s.Error("%s", err)
				continue
			}
			conn, err := mocknet.Dial("tcp", addr, s.Localhost())
			if err != nil {
				continue
			}
			readyCh, pChan, _ := util.ReadFletaPacket(conn)
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
				s.Error("%s", err)
				continue
			}
			<-readyCh
			conn.Write(p)

		} else {
			s.ConsignmentCast(peerlist.PeerList{}.Location(), fp)
		}

		time.Sleep(time.Second * 3)
	}
}

func (s *Sync) requestBlockFromObserver() error {
	addr := s.fi.GetObserverNodeAddr()

	conn, err := mocknet.Dial("tcp", addr, s.Localhost())
	if err != nil {
		return err
	}
	readyCh, pChan, _ := util.ReadFletaPacket(conn)
	go func() {
		fp := <-pChan
		if fp.Command == "SYRSGEBL" {
			var block Block
			util.FromJSON(&block, fp.Content)
			s.Log("%s", block)
			s.putGenesisBlock(block)
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
	<-readyCh
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
	s.RLock()
	height := s.BlockList[len(s.BlockList)-1].Height
	s.RUnlock()
	return height
}

func (s *Sync) LastedBlock() *Block {
	index := len(s.BlockList) - 1
	return s.BlockList[index]
}

func (s *Sync) PushBlock(block *Block) bool {
	if block.Height == 0 {
		s.pushBlock(block)
		return true
	}
	if bb := s.BlockList[block.Height-1]; bb != nil {
		if cb := s.BlockList[block.Height]; cb == nil {
			s.pushBlock(block)
			return true
		}
	}
	return false
}

func (s *Sync) MakeBlock(node *flanetinterface.Node) error {
	length := len(s.BlockList)
	if length == 0 {
		return ErrNotFoundGenesisBlock
	}
	prevBlock := s.BlockList[length-1]
	if prevBlock.Height == length-1 {
		block := &Block{
			MakeBlockTime: time.Now(),
			Addr:          node.Addr(),
			Height:        length,
		}

		s.pushBlock(block)
		node.BlockTime = block.MakeBlockTime
		//  = block.MakeBlockTime
		return nil
	}

	return ErrInvalidHeight
}

//CheckBlock TODO check other block height
func (s *Sync) pushBlock(block *Block) {
	if block.Height == 0 {
		s.BlockList = []*Block{block}
	} else {
		s.BlockList = append(s.BlockList, block)
	}
	s.blockAddrMap[block.Addr] = block

	err := s.fi.NewBlock(block)
	if err != nil {
		s.Error("%s", err)
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
	PeerSend(string, util.FletaPacket) error
	GetObserverNodeAddr() string
	PeerList() ([]flanetinterface.Node, error)
	NewBlock(*Block) error
	SeedNodeAddr() (string, error)
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
	if s.Localhost() == seedNodeAddr {
		return "", ErrNotFoundSeedNode
	}
	return seedNodeAddr, nil
}

//GetBlockHeight return height of block
func (s *Sync) GetBlockHeight() int {
	return len(s.BlockList)
}
