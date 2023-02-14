package tutorial

import "sync"

var wt sync.WaitGroup

func Testsync() {
	wt.Add(10)
	for i := 0; i < 10; i++ {
		Mutex(i)
	}
	wt.Wait()
}
func Testgeneric() {
	generic(true)
	generic("hello")
	generic(67)
}
func TestDecode() {
	decodexml("config.xml")
}
