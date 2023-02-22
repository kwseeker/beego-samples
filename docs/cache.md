# Beego Cache

源码包：github.com/astaxie/beego/cache

```shell
cache
├── cache.go			# Cache接口定义；var adapters = make(map[string]Instance) 缓存实现策略集合
├── cache_test.go		
├── conv.go				# 上面的查询接口返回的都是 interface{}, 这里定义了将结果转换成其他类型的函数
├── conv_test.go
├── file.go				# 使用文件实现的缓存
├── memcache			
│   ├── memcache.go		# memcache做缓存
│   └── memcache_test.go
├── memory.go			# 内存作为缓存
├── README.md
├── redis
│   ├── redis.go		# redis（redigo）作为缓存
│   └── redis_test.go
└── ssdb
    ├── ssdb.go			# ssdb作为缓存
    └── ssdb_test.go
```



## 内存作缓存

原理：核心是 map[string]*interface{} + 读写锁(sync.RWMutex) + vacuum() 定时清除超时的键值的协程。

```go
type MemoryItem struct {
	val         interface{}
	createdTime time.Time
	lifespan    time.Duration
}
type MemoryCache struct {
	sync.RWMutex			//以组合方式继承读写锁
	dur   time.Duration
	items map[string]*MemoryItem
	Every int // run an expiration check Every clock time
}
```



## Redis作缓存

每次执行 NewCache() 都会创建一个连接池。如果想复用连接池可以紧挨嗯NewCache()返回的对象存储起来。

```go
//用于创建连接的Redis配置，NewCache的返回值
type Cache struct {
	p        *redis.Pool 	// redis连接池
	conninfo string			//
	dbNum    int			// 数据库num
	key      string			// 其实是固定的前缀
	password string
	maxIdle  int
	//the timeout to a value less than the redis server's timeout.
	timeout time.Duration
}

//init方法中只是注册了实例化Redis配置的函数，StartAndGC()方法中真正会创建连接并返回连接实例
func init() {
	cache.Register("redis", NewRedisCache)
}
```

但是StartAndGC()看起来有点诡异（怎么刚获取的连接在退出方法时就关闭了），研究下内部实现：

```go
func (rc *Cache) StartAndGC(config string) error {
    ...
    rc.connectInit()
    c := rc.p.Get()
    defer c.Close()
    ...
}
```

