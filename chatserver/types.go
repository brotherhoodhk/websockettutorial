package chatserver

import "github.com/gorilla/websocket"

var conpool = make(map[*Connection]struct{})
var roompool = make(map[string][]*Connection)
var customerpool = make(map[*Connection]*userinfo)
var dishespool = make(map[*singledish]bool) //菜单池，false表示点单未出单，true表示出单
var ordhub = &OrderHub{SendToChef: make(chan *singledish), DishDone: make(chan *singledishplus)}
var chefpool = make(map[*Connection]struct{})
var dishlink = make(map[int]*singledish) //订单id与具体订单绑定
var deskinfo = make(map[int]*userinfo)   //将订单讯息与餐位绑定
type basicmsg struct {
	Content string `json:"content"`
	Sign    string `json:"sign"`
}
type Connection struct {
	con  *websocket.Conn
	send chan []byte
}
type Hub struct {
	register      chan *Connection
	unregister    chan *Connection
	customer      chan *Connection
	clearcustomer chan *Connection
	Broadcast     chan []byte
}

// order system
type order struct {
	Header string `json:"header"`
	Dish   []int  `json:"dish"`
}

// 记录点菜的客户讯息
type userinfo struct {
	Id      int     `json:"id"`      //用于识别客户的id
	OrderId int     `json:"orderid"` //客户点菜的菜单id，一个客户可以持有多份菜单，一个菜单可持有一份以上的菜，结算以每份菜单具体金额为准
	Sum     float32 `json:"sum"`     //客户一共需要结算的金额
	Status  bool    `json:"status"`  //客户是否买单
	Action  string  `json:"action"`  //记录客户的所有点菜操作
}
type singledish struct {
	Name string `json:"name"`
	Time string `json:"time"`
	Id   int    `json:"id"`
}
type aboutdish struct {
	Name  string
	Price float32
	Id    int
}
type OrderHub struct {
	SendToChef chan *singledish
	DishDone   chan *singledishplus
}

// singledish加强版，用于接受厨师端的讯息
type singledishplus struct {
	Dishinfo *singledish `json:"dishinfo"`
	Orderid  int         `json:"orderid"` //订单唯一的id，由5位10进制数组成，用于识别客户点的每一道菜
}
