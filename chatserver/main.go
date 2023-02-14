package chatserver

import (
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"newstart/server"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
)

var upgrade = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
var hub = &Hub{register: make(chan *Connection), unregister: make(chan *Connection), Broadcast: make(chan []byte, 512), customer: make(chan *Connection), clearcustomer: make(chan *Connection)}

func (s *Hub) Run() {
	fmt.Println("start register machine")
	for {
		rand.Seed(time.Now().UnixNano())
		select {
		case c := <-s.register:
			conpool[c] = struct{}{}
			fmt.Println("register connection")
		case c := <-s.unregister:
			if _, ok := conpool[c]; ok {
				delete(conpool, c)
				fmt.Println("unregister connection")
			}
		//order system zone start
		case c := <-s.customer:
			conpool[c] = struct{}{}
			id := rand.Intn(89999) + 10000
			orderid := rand.Intn(899999) + 100000
			customerpool[c] = &userinfo{Id: id, OrderId: orderid, Status: false, Sum: 0, Action: ""}
		case c := <-s.clearcustomer:
			if _, ok := conpool[c]; ok {
				delete(conpool, c)
			}
			if _, ok := customerpool[c]; ok {
				delete(customerpool, c)
			}
		case c := <-ordhub.SendToChef:
			for con, _ := range chefpool {
				err := con.con.WriteJSON(c)
				errorlog.Println(err)
			}
		//order system zone end
		case m := <-s.Broadcast:
			for clients := range conpool {
				select {
				case clients.send <- m:
				default:
					delete(conpool, clients)
					fmt.Println("delete connection")
				}
			}
		}
	}
}
func ServerStart() {
	go hub.Run()
	http.HandleFunc("/chat", ChatRoom)
	http.HandleFunc("/orderdish", OrderSomething)
	http.ListenAndServe(":8001", nil)
}
func ChatRoom(w http.ResponseWriter, r *http.Request) {
	upgrade.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	heads := r.URL.Query()
	if !server.CheckArgs([]string{"roomid", "usrid"}, heads) {
		fmt.Println("args not enough")
		return
	}
	roomid := heads["roomid"][0]
	// usrid:=heads["usrid"][0]
	con := &Connection{con: ws, send: make(chan []byte, 512)}
	if _, ok := roompool[roomid]; !ok {
		roompool[roomid] = []*Connection{}
	}
	conlist := roompool[roomid]
	conlist = append(conlist, con)
	roompool[roomid] = conlist
	hub.register <- con
	defer func() {
		hub.unregister <- con
		con.con.Close()
	}()
	for {
		var recmsg basicmsg
		err = con.con.ReadJSON(&recmsg)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("rec msg:%+v\n", recmsg)
		recmsg.Sign = "from server"
		for _, wscon := range conlist {
			err = wscon.con.WriteJSON(&recmsg)
			fmt.Println(recmsg)
			if err != nil {
				fmt.Println(err)
			}
		}
		hub.Broadcast <- []byte(recmsg.Sign)
	}
}

var errorlog = LogInit("errorlog")

// ordering system
func OrderSomething(w http.ResponseWriter, r *http.Request) {
	upgrade.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		errorlog.Println(err)
		return
	}
	con := &Connection{con: ws, send: make(chan []byte, 256)}
	hub.register <- con
	defer func() {
		hub.unregister <- con
		ws.Close()
	}()
	Waiter(con)
}
func Waiter(con *Connection) {
	var orderinfo order
	for {
		err := con.con.ReadJSON(&orderinfo)
		if err != nil {
			errorlog.Println("read from connection failed,", err)
		} else {
			userorderinfo := customerpool[con]
			price, actionid := getdishinfo(orderinfo.Dish)
			userorderinfo.Sum += price
			userorderinfo.Action += actionid + "\n"
			//这里还应有一个推送功能，用于将新点菜品发送给后厨端
			dish := new(singledish)
			dish.ScanAction(actionid)
			dish.PushToPool()
			ordhub.SendToChef <- dish
		}
	}
}
func getdishinfo(id int) (float32, string) {
	//从数据库中获取菜品讯息
	dbcon, err := sql.Open("mysql", "test:123456@tcp(127.0.0.1)/lab?charset=utf8")
	if err != nil {
		errorlog.Println(err)
		return -1, ""
	}
	defer dbcon.Close()
	var dishinfo aboutdish
	dishinfo.Id = id
	dbcon.QueryRow("select price,name from ordersystem where id =?", id).Scan(&dishinfo.Price, &dishinfo.Name)
	if err != nil {
		errorlog.Println(err)
	}
	//操作id=下单时间+菜品名称+菜品id
	actionid := time.Now().Format(time.Kitchen) + "&" + dishinfo.Name + "&" + strconv.Itoa(dishinfo.Id)
	return dishinfo.Price, actionid
}
