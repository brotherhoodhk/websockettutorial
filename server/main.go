package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var errorlog = loginit("error")
var upgrade = websocket.Upgrader{ReadBufferSize: 10 * MB, WriteBufferSize: 10 * MB}
var hub = Hub{Register: make(chan *User), UnRegister: make(chan *User)}

func ServerStart() {
	go hub.Run()
	http.HandleFunc("/chat", chatroom)
	http.HandleFunc("/greet", Greeting)
	errorlog.Println(http.ListenAndServe(":8001", nil))
}
func chatroom(w http.ResponseWriter, r *http.Request) {
	fmt.Println("get connection.../")
	upgrade.CheckOrigin = func(r *http.Request) bool { return true }
	con, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		errorlog.Println(err)
	}
	// roomid := r.URL.Query().Get("roomid")
	// userid := r.URL.Query().Get("usrid")
	heads := r.URL.Query()
	if !CheckArgs([]string{"roomid", "usrid"}, heads) {
		fmt.Println("args are not enough")
		return
	}
	roomid := heads["roomid"][0]
	userid := heads["usrid"][0]
	user := User{Con: con, Roomid: roomid, Userid: userid}
	hub.Register <- &user
	defer func() {
		hub.UnRegister <- &user
	}()
	Reader(&user)
}
func Greeting(w http.ResponseWriter, r *http.Request) {
	upgrade.CheckOrigin = func(r *http.Request) bool { return true }
	con, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	var basicmsg = BasicMessage{Roomid: "hello", Userid: "cho", Content: "hello world", Sign: "from server"}
	var recmsg BasicMessage
	for {
		err = con.ReadJSON(&recmsg)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = con.WriteJSON(&basicmsg)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
