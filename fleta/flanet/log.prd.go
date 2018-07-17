// +build !debugoption

package flanet

import (
	"fmt"
	"log"
	"runtime"
	"strings"

	util "fleta/samutil"
)

//Log It's log
func (f *Flanet) Log(format string, msg ...interface{}) {
	msg = append([]interface{}{f.getFlanetID(), util.Sha256HexInt(f.getFlanetID())}, msg...)

	format = string(append([]byte("%d, %s "), []byte(format)...))
	log.Printf(format, msg...)
}

//Error error log
func (f *Flanet) Error(format string, msg ...interface{}) {
	msg = append([]interface{}{f.getFlanetID(), util.Sha256HexInt(f.getFlanetID())}, msg...)

	buf := make([]byte, 1<<16)
	runtime.Stack(buf, true)
	strs := strings.Split(string(buf), "\n\n")
	msg = append(msg, []interface{}{strs[0]}...)

	format = fmt.Sprintf("%s %s\n%s", "Error %d, %s ", format, "%s")

	// log.Printf(format, msg...)
}

//Debug debug log
func (f *Flanet) Debug(format string, msg ...interface{}) {
}

//Info info log
func (f *Flanet) Info(format string, msg ...interface{}) {
}
