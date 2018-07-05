// +build debugoption

package flanet

import (
	"fleta/util"
	"log"
)

func (f *Flanet) Log(format string, msg ...interface{}) {
	msg = append([]interface{}{f.getFlanetID(), util.Sha256HexInt(f.getFlanetID())}, msg...)

	format = string(append([]byte("%d, %s log"), []byte(format)...))
	log.Printf(format, msg...)
}

func (f *Flanet) Debug(format string, msg ...interface{}) {
	msg = append([]interface{}{f.getFlanetID(), util.Sha256HexInt(f.getFlanetID())}, msg...)

	format = string(append([]byte("%d, %s "), []byte(format)...))
	log.Printf(format, msg...)
}

func (f *Flanet) Info(format string, msg ...interface{}) {
	msg = append([]interface{}{f.getFlanetID(), util.Sha256HexInt(f.getFlanetID())}, msg...)

	format = string(append([]byte("%d, %s "), []byte(format)...))
	log.Printf(format, msg...)
}
