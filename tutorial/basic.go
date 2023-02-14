package tutorial

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
)

func WR(filename string) {
	var buffer = make([]byte, 1024)
	//use ioutil
	f, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("read file failed,", err)
	}
	fmt.Println(string(f))
	//use os openfile
	fe, err := os.OpenFile(filename, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println("read file failed,", err)
	}
	n, err := fe.Read(buffer)
	if err != nil {
		fmt.Println("read file failed,", err)
	}
	fmt.Println(string(buffer[:n]))
}
func WRMutex(filename string) {
	var wrmutex sync.RWMutex
	//write lock
	wrmutex.Lock()
	err := ioutil.WriteFile(filename, []byte("i write it"), 0666)
	if err != nil {
		fmt.Println(err)
	}
	//write unlock
	wrmutex.Unlock()
	//read lock
	wrmutex.RLock()
	f, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("read from file:%v\n", string(f))
	//read unlock
	wrmutex.RUnlock()
}

var mutex sync.Mutex

func Mutex(num int) {
	mutex.Lock()
	fmt.Println("do it ", num)
	mutex.Unlock()
	wt.Done()

}

type speicalstruct interface{ int | string | bool | byte }

func generic[T speicalstruct](data T) {
	genericmap := make(map[T]bool)
	genericmap[data] = true
	fmt.Println(genericmap)
}

type car interface {
	speed() int
	getwheels() int
	getname() string
}
type sportcar struct {
	maxspeed int
	wheels   int
	name     string
	people   []string
}
type truck struct {
	wheels   int
	name     string
	maxspeed int
}

func (s *sportcar) speed() int {
	return s.maxspeed
}
func (s *truck) speed() int {
	return s.maxspeed
}
func (s *sportcar) getwheels() int {
	return s.wheels
}
func (s *truck) getwheels() int {
	return s.wheels
}
func (s *sportcar) getname() string {
	return s.name
}
func (s *truck) getname() string {
	return s.name
}
func info(s car) {
	fmt.Printf("maxspeed:%v,wheels:%v,driver name:%v\n", s.speed(), s.getwheels(), s.getname())
}

type config struct {
	Name    xml.Name `xml:"server"`
	address string
	port    int
	authkey string
}

func decodexml(filename string) {
	configbase := &config{address: "localhost", port: 3000, authkey: "hellocho", Name: xml.Name{Local: "server"}}
	fe, err := xml.Marshal(configbase)
	if err != nil {
		fmt.Println(err)
		return
	}
	ioutil.WriteFile(filename, fe, 0666)
	f, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
	}
	var newconfig config
	err = xml.Unmarshal(f, &newconfig)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("config info:%+v\n", newconfig)
}
