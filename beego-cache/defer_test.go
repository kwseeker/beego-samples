package beego_cache

import (
	"log"
	"testing"
)

/*
	func (rc *Cache) StartAndGC(config string) error {
		...
		rc.connectInit()
		c := rc.p.Get()
		defer c.Close()
		...
	}
*/
type Obj struct {
}

func (o *Obj) func1() {
	log.Println("func1")
	defer log.Println("defer in func1")
}

func (o *Obj) func2() {
	log.Println("func1")
}

func TestDeferFunc(t *testing.T) {
	obj := &Obj{}
	obj.func1()
	obj.func2()
}
