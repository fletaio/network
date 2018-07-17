package network

import (
	"log"
	"strconv"

	"fleta/flanetinterface"
	"fleta/mock"
	util "fleta/samutil"
)

var fmList []string

func init() {
	totalCount := simulationdata.NormalNodeStartIndex + simulationdata.NormalNodeCount
	for i := 0; i < totalCount; i++ {
		if itoNodeType(i) == flanetinterface.FormulatorNode {
			fmAddr := util.Sha256HexInt(i)
			fmList = append(fmList, fmAddr)
		}
	}
}

func mockDataSend(index int) {
	// formulatorListSend(index)

	// peerlistSend(index)
}

func formulatorListSend(index int) {
	localhost := util.Sha256HexInt(index)
	cRead, cWrite := RegistDial("tcp", localhost, localhost)

	log.Println("len :", len(fmList))
	seri := util.ToJSON(fmList)

	fp := util.FletaPacket{
		Command:     "FMHSSEND",
		Compression: false,
		Content:     seri,
	}

	if packet, err := fp.Packet(); err != nil {
		log.Fatal(err)
	} else {
		cWrite.Write(packet)
	}

	if err := cWrite.Close(); err != nil {
		log.Fatal(err)
	}
	if err := cRead.Close(); err != nil {
		log.Fatal(err)
	}

}

func peerlistSend(index int) {
	nodes := make([]flanetinterface.Node, 0)
	totalCount := simulationdata.ObserverNodeCount
	i := 0
	for ; i < totalCount; i++ {
		node := flanetinterface.Node{
			Address:  util.Sha256HexString(strconv.Itoa(i)),
			NodeType: flanetinterface.ObserverNode,
		}
		nodes = append(nodes, node)
	}
	totalCount += simulationdata.FormulatorNodeCount
	for ; i < totalCount; i++ {
		node := flanetinterface.Node{
			Address:  util.Sha256HexString(strconv.Itoa(i)),
			NodeType: flanetinterface.FormulatorNode,
		}
		nodes = append(nodes, node)
	}
	totalCount += simulationdata.NormalNodeCount
	for ; i < totalCount; i++ {
		node := flanetinterface.Node{
			Address:  util.Sha256HexString(strconv.Itoa(i)),
			NodeType: flanetinterface.NormalNode,
		}
		nodes = append(nodes, node)
	}

	seri := util.ToJSON(nodes)

	fp := util.FletaPacket{
		Command:     "PLRTPELT",
		Compression: false,
		Content:     seri,
	}

	packet, err := fp.Packet()
	if err != nil {
		log.Println("err ", err)
	}
	localhost := util.Sha256HexInt(index)
	cRead, cWrite := RegistDial("tcp", localhost, localhost)

	cWrite.Write(packet)

	cWrite.Close()
	cRead.Close()
}
