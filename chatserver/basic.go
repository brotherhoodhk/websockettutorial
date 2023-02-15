package chatserver

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	ROOTPATH = "/Users/oswaldo/dev/golang/newstart"
	LOGPATH  = ROOTPATH + "/logs/"
)

var errorlog = LogInit("errorlog")
var processlog = LogInit("process")

func LogInit(logname string) *log.Logger {
	f, err := os.OpenFile(LOGPATH+logname+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println("open file failed,", err)
		return nil
	}
	newlog := log.New(f, "["+logname+"]", log.LUTC|log.Lshortfile|log.LstdFlags)
	return newlog
}
func (s *singledish) ScanAction(action string) {
	acarray := strings.Split(action, "&")
	s.Name = acarray[1]
	s.Time = acarray[0]
	var err error
	s.Id, err = strconv.Atoi(acarray[2])
	if err != nil {
		errorlog.Println(err)
	}
}

// 将此操作添加到后厨菜单池
func (s *singledish) PushToPool() {
	dishespool[s] = false
}
