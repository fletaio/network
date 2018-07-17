package formulator

import (
	"net"

	util "fleta/samutil"
)

func (fm *Formulator) addProcessCommand() {
	fm.AddCommand("FMHSRQFL", func(conn net.Conn, fp util.FletaPacket) (exit bool, err error) {
		fl, err := fm.hardstate.FormulatorAddrList()
		if err != nil {
			return false, err
		}
		sendFp := util.FletaPacket{
			Command:     "FMHSSEND",
			Compression: true,
			Content:     util.ToJSON(fl),
		}

		p, err := sendFp.Packet()
		if err != nil {
			return false, err
		}
		conn.Write(p)
		return false, nil
	})
	fm.AddCommand("FMHSSEND", func(conn net.Conn, fp util.FletaPacket) (exit bool, err error) {
		var nodes []string
		util.FromJSON(&nodes, fp.Content)
		fm.hardstate.AddCandidateNodeAddr(nodes)
		return false, nil
	})
	fm.AddCommand("FMHDRQFM", func(conn net.Conn, fp util.FletaPacket) (exit bool, err error) {
		fp = util.FletaPacket{
			Command: "FMHDRSFM",
			Content: fm.GetHint(),
		}
		p, err := fp.Packet()
		if err != nil {
			return false, err
		}
		conn.Write(p)
		return false, nil
	})
}
