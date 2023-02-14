package server

import "github.com/gorilla/websocket"

const (
	KB = 1024
	MB = 1024 * KB
	GB = 1024 * MB
)

type User struct {
	Con    *websocket.Conn
	Roomid string
	Userid string
}
type Hub struct {
	Register   chan *User
	UnRegister chan *User
	Broadcast  chan []byte
}
type BasicMessage struct {
	Roomid  string `json:"roomid"`
	Userid  string `json:"usrid"`
	Content string `json:"content"`
	Sign    string `json:"sign"`
}
type Conn struct {
	Con    *websocket.Conn
	Buffer chan []byte
}

var RoomPool = make(map[string][]*User)
var ConnectionPool = make(map[*Conn]struct{})

type SpecialStruct interface{ *User | string | []string }
