// +build !debugoption

package flanet

import (
	"fleta/util"
	"log"
)

func (f *Flanet) Log(format string, msg ...interface{}) {
	// if f.GetFlanetID() == 0 || f.GetFlanetID() == 1 {
	msg = append([]interface{}{f.getFlanetID(), util.Sha256HexInt(f.getFlanetID())}, msg...)

	format = string(append([]byte("%d, %s "), []byte(format)...))
	log.Printf(format, msg...)
	// }
}

func (f *Flanet) Debug(format string, msg ...interface{}) {
}

func (f *Flanet) Info(format string, msg ...interface{}) {
}
