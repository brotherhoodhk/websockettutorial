package chatserver

import (
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	ROOTPATH = "/Users/oswaldo/dev/golang/newstart"
	LOGPATH  = ROOTPATH + "/logs/"
)

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
	s.time = acarray[0]
}

// 将此操作添加到后厨菜单池
func (s *singledish) PushToPool() {
	dishespool[s] = false
}
