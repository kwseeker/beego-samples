package beego_cache

import (
	"github.com/astaxie/beego/cache"
	_ "github.com/astaxie/beego/cache/redis"
	"testing"
	"time"
)

func TestRedisCache(t *testing.T) {
	//config中key是前缀
	//bm内部有个连接池
	bm, err := cache.NewCache("redis", `{"key":"rc","conn":":6379","dbNum":"0","password":""}`)
	if err != nil {
		t.Fatal("err: ", err)
	}

	_ = bm.Put("k", "v", time.Second*10)
	v1 := bm.Get("k")
	t.Log("v1: ", v1)
	v2 := bm.Get("key1")
	t.Logf("v2: %v\n", v2)
}
