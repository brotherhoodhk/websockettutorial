package server

import (
	"fmt"
	"log"
	"os"
)

func loginit(logname string) *log.Logger {
	f, err := os.OpenFile(logname+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(err)
	}
	newlogger := log.New(f, "["+logname+"]", log.LUTC|log.Lshortfile|log.LstdFlags)
	return newlogger
}
func CheckArgs[T SpecialStruct](args []string, target map[string]T) bool {
	for _, v := range args {
		if _, ok := target[v]; !ok {
			return false
		}
	}
	return true
}
func checkerror(err error) bool {
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}
func (s *Hub) Run() {
	fmt.Println("Register machine start .../")
	for {
		select {
		case c := <-s.Register:
			RegisterRoom(c.Roomid, c)
		case m := <-s.UnRegister:
			UnRegisterRoom(m.Roomid, m)
		case m := <-s.Broadcast:
			BroadCast(m)
		}
	}
}
func RegisterRoom(roomid string, user *User) {
	//debug line
	fmt.Println("start register connection.../")
	//end
	if _, ok := RoomPool[roomid]; !ok {
		RoomPool[roomid] = []*User{}
	}
	registercon := &Conn{Con: user.Con, Buffer: make(chan []byte, 1*MB)}
	ConnectionPool[registercon] = struct{}{}
	// usrlist := RoomPool[roomid]
	// usrlist = append(usrlist, user)
	// RoomPool[roomid] = usrlist
	RoomPool[roomid] = append(RoomPool[roomid], user)
	fmt.Printf("room pool update %+v\n", RoomPool)
}
func UnRegisterRoom(roomid string, user *User) {
	//debug line
	fmt.Println("start unregister connection.../")
	//end
	if _, ok := RoomPool[roomid]; ok {
		if len(RoomPool[roomid]) == 0 {
			delete(RoomPool, roomid)
		} else {
			for k, v := range RoomPool[roomid] {
				if v == user {
					userlist := RoomPool[roomid]
					if k != len(RoomPool[roomid])-1 {
						userlist = append(userlist[:k], userlist[k+1:]...)
					} else {
						userlist = userlist[:k]
					}
					RoomPool[roomid] = userlist
					break
				}
			}
		}
	}
	user.Con.Close()
}
func BroadCast(value []byte) {
	for k := range ConnectionPool {
		select {
		case k.Buffer <- value:
		default:
			delete(ConnectionPool, k)
			close(k.Buffer)
		}
	}
}
func Reader(user *User) {
	var msg BasicMessage
	for {
		err := user.Con.ReadJSON(&msg)
		checkerror(err)
		fmt.Println(msg)
		WriteToRoom(user.Roomid, &msg)
		hub.Broadcast <- []byte(msg.Content)
	}
}
func WriteToRoom(roomid string, msg *BasicMessage) {
	if _, ok := RoomPool[roomid]; !ok {
		return
	}
	for _, user := range RoomPool[roomid] {
		err := user.Con.WriteJSON(msg)
		checkerror(err)
	}
}
