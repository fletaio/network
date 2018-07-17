// +build debugoption

package flanet

import (
	"log"

	util "fleta/samutil"
)

//Log log
func (f *Flanet) Log(format string, msg ...interface{}) {
	msg = append([]interface{}{f.getFlanetID(), util.Sha256HexInt(f.getFlanetID())}, msg...)

	format = string(append([]byte("%d, %s "), []byte(format)...))
	log.Printf(format, msg...)
}

//Error error log
func (f *Flanet) Error(format string, msg ...interface{}) {
	msg = append([]interface{}{f.getFlanetID(), util.Sha256HexInt(f.getFlanetID())}, msg...)

	format = string(append([]byte("Error %d, %s "), []byte(format)...))
	log.Printf(format, msg...)
}

//Debug debug log
func (f *Flanet) Debug(format string, msg ...interface{}) {
	msg = append([]interface{}{f.getFlanetID(), util.Sha256HexInt(f.getFlanetID())}, msg...)

	format = string(append([]byte("Debug %d, %s "), []byte(format)...))
	log.Printf(format, msg...)
}

//Info info log
func (f *Flanet) Info(format string, msg ...interface{}) {
	msg = append([]interface{}{f.getFlanetID(), util.Sha256HexInt(f.getFlanetID())}, msg...)

	format = string(append([]byte("Info %d, %s "), []byte(format)...))
	log.Printf(format, msg...)
}
