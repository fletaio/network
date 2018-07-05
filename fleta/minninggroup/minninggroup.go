package minninggroup

import (
	"net"
	"time"

	"fleta/mock/mockblock"
	"fleta/mock/mocknet"
	"fleta/util"
	"fleta/util/netcaster"
)

const MinningGroupCount = 30

type MinningGroup struct {
	mi             MinningGroupImpl
	imMinningGroup bool
	myScore        int
	GroupList      []*mockblock.BlockGen
	conns          []net.Conn
	netcaster.NetCaster
}

//Location TODO
func Location() string {
	return "MG"
}

//Location TODO
func (mg *MinningGroup) Location() string {
	return Location()
}

//New TODO
func New(mi MinningGroupImpl, cr *netcaster.CastRouter, hint string) *MinningGroup {
	mg := &MinningGroup{}
	mg.mi = mi

	mg.conns = make([]net.Conn, MinningGroupCount)

	mg.NetCaster = netcaster.NetCaster{
		mg, cr, hint,
	}

	return mg
}

func (mg *MinningGroup) mashConnect() {
	localhost := mg.Localhost()
	mg.myScore = mg.getScore(localhost)
	if mg.myScore == -1 {
		return
	}
	mg.conns[mg.myScore] = nil
	for index := mg.myScore + 1; index < mg.myScore+(MinningGroupCount/2); index++ {
		i := (index + MinningGroupCount) % MinningGroupCount
		conn, err := mocknet.Dial("tcp", mg.GroupList[i].Addr, mg.Localhost())
		if err == nil {
			fp := util.FletaPacket{
				Command: "MGEXLOOP",
			}
			p, err := fp.Packet()
			if err == nil {
				conn.Write(p)
				mg.SetConn(conn)
			}
		}
	}
}

func (mg *MinningGroup) SetConn(conn net.Conn) {
	index := mg.getScore(conn.RemoteAddr().String())
	if index >= 0 && index < MinningGroupCount {
		mg.conns[index] = conn
	}

}

func (mg *MinningGroup) getScore(addr string) int {
	gLen := len(mg.GroupList)
	for i := 0; i < gLen; i++ {
		if mg.GroupList[i].Addr == addr {
			return i
		}
	}
	return -1
}

func (mg *MinningGroup) start() {
	for {
		mg.mashConnect()
		if mg.myScore == 0 {
			mockblock.MakeBlock(mg.Localhost())
			mg.imMinningGroup = false
		}
		// mg.Log("%s", mg.connsString())
		time.Sleep(time.Second * 5)
	}
}

//connsString is handling process packet
func (mg *MinningGroup) connsString() []string {
	var connstr []string
	for i := 0; i < MinningGroupCount; i++ {
		if mg.conns[i] != nil {
			connstr = append(connstr, mg.conns[i].RemoteAddr().String())
		} else {
			connstr = append(connstr, "empty")
		}
	}
	return connstr
}

//ProcessPacket is handling process packet
func (mg *MinningGroup) ProcessPacket(conn net.Conn, p util.FletaPacket) error {
	switch p.Command {
	case "MGEXLOOP":
		mg.SetConn(conn)
	}
	return nil
}

//GetConnList GetConnList
func (mg *MinningGroup) GetConnList() []net.Conn {
	return mg.conns
}

type MinningGroupImpl interface {
	CalculateScore() []*mockblock.BlockGen
}

type IMinningGroup interface {
	ImMinningGroup([]*mockblock.BlockGen)
}

//ImMinningGroup TODO
func (mg *MinningGroup) ImMinningGroup() {
	if mg.imMinningGroup == false {
		mg.imMinningGroup = true
		mg.start()
	}

}
