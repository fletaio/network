package peer

import (
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
	PeerList        *sync.Map
	PeerSize        int
	LimitPeerSize   int
	MessageResolver message.Resolver
}

//New TODO
func NewManager(resolver message.Resolver) *PeerManager {
	return &PeerManager{
		PeerList:        &sync.Map{},
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
	for pm.PeerSize > pm.LimitPeerSize || pm.PeerSize > maxPeerSize {
		pm.PeerList.Range(func(k, v interface{}) bool {
			if rand.Intn(pm.PeerSize) == 1 {
				pm.deletePeer(k.(NodeID))
				return !(pm.PeerSize > pm.LimitPeerSize || pm.PeerSize > maxPeerSize)
			}
			return true
		})
	}
	pm.PeerSize++
	pm.PeerList.Store(id, peer)
}
func (pm *PeerManager) deletePeer(id NodeID) {
	pm.PeerSize--
	pm.PeerList.Delete(id)
}

func (pm *PeerManager) AddPeer(id NodeID, conn net.Conn) error {
	if _, ok := pm.PeerList.Load(id); ok {
		pm.Delete(id)
	}
	peer := NewPeer(id, pm.MessageResolver)
	errChan := peer.Bond(conn)
	go func(id NodeID) {
		err := <-errChan
		log.Printf("err %s : %s \n", id, err)
		pm.deletePeer(id)
	}(id)
	pm.storePeer(id, peer)
	// log.Println("Peer Added: ", id)
	return nil
}

func (pm *PeerManager) GetPeer(id NodeID) (IPeer, error) {
	if v, ok := pm.PeerList.Load(id); ok {
		return v.(IPeer), nil
	}
	return nil, ErrPeerNotExist
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
	if peer, err := pm.GetPeer(id); err != nil && peer != nil {
		if conn := peer.GetConn(); conn != nil {
			conn.Close()
		}
	}
	pm.deletePeer(id)
}

func (pm *PeerManager) PeerAddrList() []string {
	sended := make([]string, 0)
	pm.PeerList.Range(func(k, v interface{}) bool {
		sended = append(sended, string(k.(NodeID)))
		return true
	})
	return sended
}
func (pm *PeerManager) Broadcast(payload *packet.Payload) []NodeID {
	sended := make([]NodeID, 0)
	pm.PeerList.Range(func(k, v interface{}) bool {
		if !v.(IPeer).CheckConn() {
			pm.deletePeer(k.(NodeID))
		} else {
			pm.Send(k.(NodeID), payload, packet.UNCOMPRESSED)
			sended = append(sended, k.(NodeID))
		}
		return true
	})
	return sended
}
