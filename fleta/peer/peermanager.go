package peer

import (
	"io"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

	"fleta/message"
	"fleta/packet"
)

var maxPeerSize int = 72

type PeerManager struct {
	// TODO: Add peer list
	// PeerList        *sync.Map
	PeerMap         map[NodeID]*Peer
	peerMapLock     sync.Mutex
	seedID          NodeID
	LimitPeerSize   int
	MessageResolver message.Resolver
}

//New TODO
func NewManager(resolver message.Resolver) *PeerManager {
	return &PeerManager{
		// PeerList:        &sync.Map{},
		PeerMap:         make(map[NodeID]*Peer),
		MessageResolver: resolver,
		LimitPeerSize:   15,
	}
}

func (pm *PeerManager) SetlimitPeerSize(limitSize int) {
	if pm.LimitPeerSize < limitSize {
		pm.LimitPeerSize = limitSize
	}
}

func (pm *PeerManager) storePeer(id NodeID, peer *Peer) {
	rand.Seed(time.Now().UTC().UnixNano())
	pm.peerMapLock.Lock()
	pLen := len(pm.PeerMap)
	if pLen == 0 {
		pm.seedID = id
	}
	pm.peerMapLock.Unlock()
	for pLen > pm.LimitPeerSize || pLen > maxPeerSize {
		index := rand.Intn(pLen)
		count := 0
		pm.peerRange(func(k NodeID, v *Peer) bool {
			if index == count && k != pm.seedID {
				pm.Delete(k)
				return false
			}
			count++
			return true
		})
		pm.peerMapLock.Lock()
		pLen = len(pm.PeerMap)
		pm.peerMapLock.Unlock()
	}

	pm.peerMapLock.Lock()
	pm.PeerMap[id] = peer
	pm.peerMapLock.Unlock()
	// pm.PeerList.Store(id, peer)
}
func (pm *PeerManager) getPeer(id NodeID) (*Peer, bool) {
	pm.peerMapLock.Lock()
	defer pm.peerMapLock.Unlock()
	// if v, ok := pm.PeerList.Load(id); ok {
	v, ok := pm.PeerMap[id]
	return v, ok
}
func (pm *PeerManager) deletePeer(id NodeID) {
	pm.peerMapLock.Lock()
	defer pm.peerMapLock.Unlock()
	delete(pm.PeerMap, id)
	// pm.PeerList.Delete(id)
}
func (pm *PeerManager) peerRange(f func(k NodeID, v *Peer) bool) {
	pm.peerMapLock.Lock()
	ks := make([]NodeID, len(pm.PeerMap))
	vs := make([]*Peer, len(pm.PeerMap))
	length := len(pm.PeerMap)
	i := 0
	for k, v := range pm.PeerMap {
		ks[i] = k
		vs[i] = v
		i++
	}
	pm.peerMapLock.Unlock()

	for i := 0; i < length; i++ {
		if !f(ks[i], vs[i]) {
			break
		}
	}
}

func (pm *PeerManager) GetPeer(id NodeID) (IPeer, error) {
	// if v, ok := pm.PeerList.Load(id); ok {
	if v, ok := pm.getPeer(id); ok {
		return v, nil
	}
	return nil, ErrPeerNotExist
}

func (pm *PeerManager) AddPeer(id NodeID, conn net.Conn) error {
	// if _, ok := pm.PeerList.Load(id); ok {
	if _, ok := pm.getPeer(id); ok {
		pm.Delete(id)
	}
	peer := NewPeer(id, pm.MessageResolver)
	errChan := peer.Bond(conn)
	go func(id NodeID) {
		err := <-errChan
		if err != io.ErrClosedPipe && err != io.EOF {
			log.Printf("err %s : %s \n", id, err)
		}
		pm.Delete(id)
	}(id)
	pm.storePeer(id, peer)
	return nil
}

func (pm *PeerManager) Send(id NodeID, payload *packet.Payload, compression packet.CompressionType) error {
	peer, err := pm.GetPeer(id)
	if err != nil {
		return err
	}
	return peer.GetConn().Send(payload, compression)
}

/*sam temp impl*/
func (pm *PeerManager) Delete(id NodeID) {
	peer, err := pm.GetPeer(id)
	pm.deletePeer(id)
	if err == nil && peer != nil {
		if conn := peer.GetConn(); conn != nil {
			conn.Close()
		}
	}
}

func (pm *PeerManager) PeerAddrList() []string {
	sended := make([]string, 0)
	pm.peerRange(func(k NodeID, v *Peer) bool {
		sended = append(sended, string(k))
		return true
	})
	return sended
}
func (pm *PeerManager) Broadcast(payload *packet.Payload) []NodeID {
	sended := make([]NodeID, 0)
	pm.peerRange(func(k NodeID, v *Peer) bool {
		if !v.CheckConn() {
			pm.Delete(k)
		} else {
			pm.Send(k, payload, packet.UNCOMPRESSED)
			sended = append(sended, k)
		}
		return true
	})
	return sended
}
