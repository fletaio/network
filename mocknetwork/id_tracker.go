package mocknetwork

import (
	"log"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

//GoTracker TODO
func GoTracker(f func(int, []interface{}), v ...interface{}) {
	go func(id int) {
		StartWithID(id, func() {
			f(id, v)
		})
	}(GetSimulationID())
}

//Tracker TODO
func Tracker(f func(int)) {
	f(GetSimulationID())
}

//StartWithID TODO
func StartWithID(i int, f func()) {
	f()
}

//GetSimulationID TODO
func GetSimulationID() int {
	buf := make([]byte, 1<<16)
	runtime.Stack(buf, true)
	strs := strings.Split(string(buf), "\n")

	re := regexp.MustCompile("git\\.fleta\\.io\\/fleta\\/mocknet\\/mocknetwork\\.StartWithID[^(]*\\(")
	for _, str := range strs {
		if strings.Contains(str, "StartWithID") {
			str = re.Split(str, 10)[1]
			str = strings.TrimRight(str, ")")
			ids := strings.Split(str, ",")
			if len(ids) > 0 {
				var num int64
				num, err := strconv.ParseInt(strings.TrimPrefix(ids[0], "0x"), 16, 32)
				if err != nil {
					log.Fatal(err)
				}
				i := int(num)
				return i
			}
		}
	}

	return -1
}
