package mocknet

import (
	"context"
	"net"
	"strings"
	"time"

	"git.fleta.io/fleta/framework/log"
)

//DialTimeout is return Conn
func DialTimeout(networkType, address string, timeout time.Duration, localhost string) (net.Conn, error) {
	connected := make(chan net.Conn)

	now := time.Now()
	earliest := now.Add(timeout)
	ctx, cancel := context.WithDeadline(context.Background(), earliest)
	defer cancel()

	delay := mockDelay(localhost, address)
	go func() {
		time.Sleep(delay)
		connected <- registDial(networkType, address, localhost)
	}()
	select {
	case <-ctx.Done():
		return nil, ErrDialTimeout
	case conn := <-connected:
		mc := &mockConn{
			Conn: conn,
			LocalAddrVal: mockAddr{
				network: networkType,
				address: localhost,
			},
			RemoteAddrVal: mockAddr{
				network: networkType,
				address: address,
			},
			targetID:      address,
			readDeadline:  -1,
			writeDeadline: -1,
		}

		return mc, nil
	}

}

//Dial is return Conn
// func Dial(networkType, address string, localhost string) (net.Conn, error) {
func Dial(network, address string) (net.Conn, error) {
	if strings.HasPrefix(network, "mock") {
		addrs := strings.Split(address, ":")
		port := addrs[len(addrs)-1]

		mockIDs := strings.Split(network, ":")
		localhost := strings.Join(mockIDs[1:], ":") + ":" + port

		var conn net.Conn

		delay := mockDelay(localhost, address)
		time.Sleep(delay)

		networkType := "mock"
		conn = registDial(networkType, address, localhost)

		mc := &mockConn{
			Conn: conn,
			LocalAddrVal: mockAddr{
				network: networkType,
				address: localhost,
			},
			RemoteAddrVal: mockAddr{
				network: networkType,
				address: address,
			},
			targetID:      address,
			readDeadline:  -1,
			writeDeadline: -1,
		}

		return mc, nil
	}

	return net.Dial(network, address)
}

// Listen announces on the local network address.
func Listen(network, addr string) (net.Listener, error) {
	if strings.HasPrefix(network, "mock") {
		addrs := strings.Split(addr, ":")
		port := addrs[len(addrs)-1]

		mockIDs := strings.Split(network, ":")
		localhost := strings.Join(mockIDs[1:], ":") + ":" + port

		log.Info("addr ", localhost)

		var l net.Listener

		ml := mockListener{
			addr: &mockAddr{
				network: "mock",
				address: localhost,
			},
		}
		log.Debug("Listen : ", localhost)

		ml.waitAccept()

		l = &ml

		return l, nil
	}

	return net.Listen(network, addr)
}

//ConnParam has Reader, Writer, network, address
type ConnParam struct {
	Conn        net.Conn
	NetworkType string
	Address     string
	DialHost    string
}

func registDial(networkType, address string, localhost string) net.Conn {
	for LoadNodeMap(address).ConnParamChan == nil {
		time.Sleep(100 * time.Millisecond)
	}

	s, c := getConnPair()

	connParam := ConnParam{
		Conn:        s,
		NetworkType: networkType,
		Address:     address,
		DialHost:    localhost,
	}
	LoadNodeMap(address).ConnParamChan <- connParam

	return c
}

func registAccept(addr string) (node NodeInfo) {
	if addr == "" {
		return NodeInfo{}
	}
	node = LoadNodeMap(addr)
	if node.Address == "" {
		node = NodeInfo{
			Address:       addr,
			ConnParamChan: make(chan ConnParam, 256),
		}
		StoreNodeMap(addr, node)
	}
	return node
}
