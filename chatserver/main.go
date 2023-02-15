package chatserver

import (
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"newstart/server"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
)

var upgrade = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
var hub = &Hub{register: make(chan *Connection), unregister: make(chan *Connection), Broadcast: make(chan []byte, 512), customer: make(chan *Connection), clearcustomer: make(chan *Connection)}
var chefregister = make(chan *Connection)
var chefunregister = make(chan *Connection)

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
		case c := <-s.clearcustomer:
			if _, ok := conpool[c]; ok {
				delete(conpool, c)
			}
			if _, ok := customerpool[c]; ok {
				delete(customerpool, c)
			}
		//order system zone end
		//chef zone start
		case c := <-chefregister:
			conpool[c] = struct{}{}
			chefpool[c] = struct{}{}
			processlog.Println("register a chef")
		case c := <-chefunregister:
			if _, ok := chefpool[c]; ok {
				delete(chefpool, c)
				processlog.Println("unregister a chef")
			}
			if _, ok := conpool[c]; ok {
				delete(conpool, c)
			}
		case c := <-ordhub.SendToChef:
			//生成唯一的订单id，并将订单id与菜品讯息绑定
			rand.Seed(time.Now().UnixNano())
			orderid := rand.Intn(89999) + 10000
			newmsg := &singledishplus{Dishinfo: c, Orderid: orderid}
			dishlink[orderid] = c
			for con, _ := range chefpool {
				err := con.con.WriteJSON(newmsg)
				if err != nil {
					errorlog.Println(err)
				}
			}
		case c := <-ordhub.DishDone:
			if dish, ok := dishlink[c.Orderid]; ok {
				dishespool[dish] = true     //菜品状态为完成
				delete(dishlink, c.Orderid) //删除菜品联系
				processlog.Println(dish.Name, " is done")
			}
		//chef zone end
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
	http.HandleFunc("/chef", ChefPlatform)
	http.HandleFunc("/checkout", CheckOut)
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
			// fmt.Println(recmsg)
			if err != nil {
				fmt.Println(err)
			}
		}
		hub.Broadcast <- []byte(recmsg.Sign)
	}
}

// ordering system
func OrderSomething(w http.ResponseWriter, r *http.Request) {
	upgrade.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		errorlog.Println(err)
	}
	//绑定桌位id
	heads := r.URL.Query()
	if _, ok := heads["desk"]; !ok {
		errorlog.Println("not desk args")
		return
	}
	deskid, err := strconv.Atoi(heads["desk"][0])
	if err != nil {
		errorlog.Println("desk id is invaild")
		return
	}
	con := &Connection{con: ws, send: make(chan []byte, 256)}
	if info, ok := deskinfo[deskid]; !ok || info.Status {
		//若此桌未使用或上一次客人已结账
		id := rand.Intn(89999) + 10000
		orderid := rand.Intn(899999) + 100000
		customerpool[con] = &userinfo{Id: id, OrderId: orderid, Sum: 0, Action: "", Status: false}
		deskinfo[deskid] = customerpool[con]
		processlog.Println("desk ", deskid, " register")
	} else if ok && !info.Status {
		customerpool[con] = deskinfo[deskid]
	}
	processlog.Println("the desk ", deskid, " info ", deskinfo[deskid])
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
			//debug zone
			if _, ok := customerpool[con]; !ok {
				fmt.Println("connection dont exsit")
				id := rand.Intn(89999) + 10000
				orderid := rand.Intn(899999) + 100000
				customerpool[con] = &userinfo{Id: id, OrderId: orderid, Sum: 0, Action: "", Status: false}
			}
			//end
			userorderinfo := customerpool[con]
			// fmt.Println(userorderinfo) //debug line
			for _, v := range orderinfo.Dish {
				price, actionid := getdishinfo(v)
				// fmt.Println(price) //debug line
				userorderinfo.Sum += price
				userorderinfo.Action += actionid + "\n"
				//这里还应有一个推送功能，用于将新点菜品发送给后厨端
				dish := new(singledish)
				dish.ScanAction(actionid)
				dish.PushToPool()
				ordhub.SendToChef <- dish
			}
		}
		con.send <- []byte("im live")
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
	err = dbcon.QueryRow("select price,name from ordersystem where id =?", id).Scan(&dishinfo.Price, &dishinfo.Name)
	if err != nil {
		errorlog.Println(err)
	}
	//操作id=下单时间+菜品名称+菜品id
	actionid := time.Now().Format(time.Kitchen) + "&" + dishinfo.Name + "&" + strconv.Itoa(dishinfo.Id)
	return dishinfo.Price, actionid
}
func ChefPlatform(w http.ResponseWriter, r *http.Request) {
	upgrade.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		errorlog.Println("establish connection with chef failed", err)
	}
	con := &Connection{con: ws, send: make(chan []byte, 256)}
	chefregister <- con
	defer func() {
		chefunregister <- con
		ws.Close()
	}()
	var dish = new(singledishplus)
	for {
		err := ws.ReadJSON(dish)
		if err == nil {
			ordhub.DishDone <- dish
		}
		//send
		con.send <- []byte("im live")
	}
}

// 结账判断
func CheckOut(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		errorlog.Println(err)
	}
	deskid, err := strconv.Atoi(r.Form.Get("deskid"))
	if err != nil {
		errorlog.Println(err)
	}
	fmt.Println("recive desk", deskid, "check out")
	if _, ok := deskinfo[deskid]; !ok {
		processlog.Println("no such desk")
		return
	}
	userfo := deskinfo[deskid]
	userfo.Status = true
	processlog.Println(deskid, " was check out")
	//将客户的点菜操作保存到磁盘
	f, err := os.OpenFile(ROOTPATH+"/data", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		errorlog.Println(err)
		return
	}
	f.Write([]byte(userfo.Action))
	f.Close()
}
