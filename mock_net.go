package network

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/fletaio/framework/log"
)

//DialTimeout is return Conn
func DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	if strings.HasPrefix(network, "mock") {
		connected := make(chan net.Conn)

		addrs := strings.Split(address, ":")
		port := addrs[len(addrs)-1]

		mockIDs := strings.Split(network, ":")
		localhost := strings.Join(mockIDs[1:], ":") + ":" + port

		now := time.Now()
		earliest := now.Add(timeout)
		ctx, cancel := context.WithDeadline(context.Background(), earliest)
		defer cancel()

		delay := mockDelay(localhost, address)
		go func() {
			time.Sleep(delay)
			conn, err := registDial(network, address, localhost)
			if err == nil {
				connected <- conn
			}
		}()
		select {
		case <-ctx.Done():
			return nil, ErrDialTimeout
		case conn := <-connected:
			mc := &mockConn{
				Conn: conn,
				LocalAddrVal: mockAddr{
					network: network,
					address: localhost,
				},
				RemoteAddrVal: mockAddr{
					network: network,
					address: address,
				},
				readDeadline:  -1,
				writeDeadline: -1,
			}

			return mc, nil
		}
	}
	return net.DialTimeout(network, address, timeout)

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
		conn, err := registDial(networkType, address, localhost)
		if err != nil {
			return nil, err
		}
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

		var l net.Listener

		ml := mockListener{
			addr: &mockAddr{
				network: "mock",
				address: localhost,
			},
		}

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

func registDial(networkType, address string, localhost string) (net.Conn, error) {
	count := 0
	for LoadNodeMap(address).ConnParamChan == nil {
		time.Sleep(100 * time.Millisecond)
		count++
		if count > 10*5 {
			return nil, ErrDialTimeout
		}
	}

	s, c, err := getConnPair()
	if err != nil {
		return nil, err
	}

	connParam := ConnParam{
		Conn:        s,
		NetworkType: networkType,
		Address:     address,
		DialHost:    localhost,
	}
	LoadNodeMap(address).ConnParamChan <- connParam

	return c, nil
}

func registAccept(addr string) (node NodeInfo) {
	log.Debug("registAccept ", addr)
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
