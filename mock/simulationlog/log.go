package simulationlog

import (
	"flag"
	"go/build"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

var logfile *os.File

var (
	Logger *log.Logger
	sTime  time.Time
)

func init() {
	// set location of log file
	var logpath = build.Default.GOPATH + "/src/info.log"

	flag.Parse()
	var file, err1 = os.Create(logpath)

	if err1 != nil {
		panic(err1)
	}
	Logger = log.New(file, "", log.LstdFlags|log.Lshortfile)
	Logger.SetFlags(0)

	sTime = time.Now()
	time := string([]byte(sTime.Format("2006-01-02T15:04:05.999999999") + "0000000000")[:30])
	Logger.Printf("t,c,d\n")
	Logger.Printf("0s,I,%s\n", time)
	addrMap = make(map[string]string)
}

var (
	addrMap     map[string]string
	addrMapLock sync.Mutex
)

func getShotcut(key string) string {
	addrMapLock.Lock()
	defer addrMapLock.Unlock()

	if v, ok := addrMap[key]; ok {
		return v
	}
	v := strconv.Itoa(len(addrMap))
	addrMap[key] = v
	return v
}

func Listen(addr string) {
	// logMsg("L,%s", getShotcut(addr))
}
func Dial(from, to string, pingTime time.Duration) {
	// logMsg("D,%s>%s:%d", getShotcut(from), getShotcut(to), pingTime)
}
func Send(from, to string) {
	// if msg, ok := msgType.(*message.MessagePayload); ok {
	// 	logMsg("S,%s>%s:%d", getShotcut(from), getShotcut(to), msg.MsgType)
	// } else {
	// logMsg("S,%s>%s:", getShotcut(from), getShotcut(to))
	// }
}
func Close(from, to string) {
	// logMsg("C,%s>%s", getShotcut(from), getShotcut(to))
}

func LogMsg(msg ...interface{}) {
	now := time.Now()
	msg = append([]interface{}{now.Sub(sTime), " "}, msg...)
	Logger.Print(msg...)
	// logMsg(format, msg...)
}

func LogMsgf(f string, msg ...interface{}) {
	now := time.Now()
	Logger.Printf(now.Sub(sTime).String()+" "+f, msg...)
	// logMsg(format, msg...)
}

func logMsg(format string, msg ...interface{}) {
	now := time.Now()
	sub := now.Sub(sTime)

	// time := string(append([]byte(time.Now().Format("2006-01-02T15:04:05.999999999")), []byte{48, 48, 48, 48, 48, 48, 48, 48, 48}...)[:30])

	msg = append([]interface{}{sub}, msg...)

	format = "%s," + format + "\n"
	// format = string(append([]byte("%0s "), append([]byte(format), []byte("\n")...)...))

	// log.Printf(format, msg...)
	Logger.Printf(format, msg...)
}
