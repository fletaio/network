// +build !debugoption

package flanet

import (
	"flag"
	"fmt"
	"go/build"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	util "fleta/samutil"
)

var logfile *os.File

var (
	Log     *log.Logger
	srcPath string
)

func init() {
	// set location of log file
	var logpath = build.Default.GOPATH + "/src/info.log"

	flag.Parse()
	var file, err1 = os.Create(logpath)

	if err1 != nil {
		panic(err1)
	}
	Log = log.New(file, "", log.LstdFlags|log.Lshortfile)
	Log.Println("LogFile : " + logpath)

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	srcPath = strings.Replace(dir, "\\", "/", 10)
}

func write(msg string) {
	if logfile == nil {
		f, err := os.OpenFile("log.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}
		logfile = f
	}

	if _, err := logfile.WriteString(msg); err != nil {
		panic(err)
	}
}

//Log It's log
func (f *Flanet) Log(format string, msg ...interface{}) {
	msg = append([]interface{}{f.getFlanetID(), util.Sha256HexInt(f.getFlanetID())}, msg...)

	buf := make([]byte, 1<<16)
	runtime.Stack(buf, true)
	strs := strings.Split(string(buf), "\n\n")
	strs = strings.Split(strs[0], "\n")
	path := strings.TrimLeft(strs[3], "\t"+srcPath)
	msg = append([]interface{}{path}, msg...)

	format = string(append([]byte("%s : %d, %s "), []byte(format)...))
	// write(fmt.Sprintf(format, msg...))
	// Log.Printf(format, msg...)
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
	// write(fmt.Sprintf(format, msg...))
	// Log.Printf(format, msg...)
	log.Printf(format, msg...)

}

//Debug debug log
func (f *Flanet) Debug(format string, msg ...interface{}) {
}

//Info info log
func (f *Flanet) Info(format string, msg ...interface{}) {
}
