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

Beego简单封装了Redigo的接口，即本质是用Redigo操作对Redis读写的。

每次执行 NewCache() 都会创建一个连接池。Redigo连接池对连接的管理逻辑和Java的Jedis类似。 如果想复用连接池可以紧挨嗯NewCache()返回的对象存储起来。

另外对Redis读写操作的方法内部封装了`从连接池获取连接操作完毕放回连接池`的逻辑。

```go
//用于创建连接的Redis配置，NewCache的返回值
type Cache struct {
	p        *redis.Pool 	// redis连接池
	conninfo string			//连接配置信息
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

但是StartAndGC()看起来有点诡异：怎么刚获取的连接在退出方法时就关闭了？

看源码实现后发现Redigo连接管理这部分其实和Java框架Jedis封装的差不多，比如Close()并不是真的就直接关闭了；
StartAndGC() 其实是初始化线程池然后新建连接再将连接放回连接池，看起来有点多余，只初始化线程池不就行了？
我猜测可能是为了：
1）提升速度：一般生产环境NewCache的对象是要复用的，经常在服务启动阶段就创建了，这样Redis读写来了直接就有一个连接可用；
2）也可以在NewCache()阶段提前测试能否成功创建连接。

内部实现：

```go
func (rc *Cache) StartAndGC(config string) error {
    ...
  	//实例化Redigo连接池（赋值给rc.p），并赋值Dial方法（建立tcp连接、依次发送AUTH、SELECT，执行认证、选择数据库等操作）
    rc.connectInit()
    //通过Redigo连接池获取连接
    //代码逻辑：
    //1）先判断连接池是否繁忙（所有连接都被占用且连接数达到约定的最大值），是的话就阻塞等待有可用连接
    //2）IdelTimeout是否大于0, 是的话就扫描一遍清除空闲超时的连接
    //3）从连接池头部获取可用的空闲连接，获取成功直接返回这个连接
    //4）获取空闲连接失败且连接数未达到最大就通过Dial()新建连接
    //注意新建连接的过程中并没有将新建的连接直接放到连接池
    c := rc.p.Get()
    //Close代码逻辑：
    //1）先释放业务代码中对连接的引用；
    //2）特殊情况处理：
    //2.1）开启事务的连接，根据状态可能要 DISCARD UNWATCH
    //2.2）开启发布订阅的连接，根据状态可能要 UNSUBSCRIBE PUNSUBSCRIBE
    //3）发一个空命令，？没看明白这步有什么用，看调用链路也没做什么有用的操作
    //4）将连接（作为空闲连接）放回线程池的头部，如果空闲连接数大于最大空闲连接数，就真正关闭，如果传参要求强制关闭也会真正关闭连接；
    //5）向 p.ch 信道中写数据，让之前还在阻塞等待获取连接的协程继续尝试获取连接
    defer c.Close()
    ...
}

//获取的连接的数据结构
type activeConn struct {
	p     *Pool					//所属连接池
	pc    *poolConn		//获取的连接对象
	state int					//连接状态
}

//Redigo线程池
type Pool struct {
	// Dial is an application supplied function for creating and configuring a
	// connection.
	//
	// The connection returned from Dial must not be in a special state
	// (subscribed to pubsub channel, transaction started, ...).
	Dial func() (Conn, error)

	// TestOnBorrow is an optional application supplied function for checking
	// the health of an idle connection before the connection is used again by
	// the application. Argument t is the time that the connection was returned
	// to the pool. If the function returns an error, then the connection is
	// closed.
	TestOnBorrow func(c Conn, t time.Time) error

	// Maximum number of idle connections in the pool.
	MaxIdle int

	// Maximum number of connections allocated by the pool at a given time.
	// When zero, there is no limit on the number of connections in the pool.
	MaxActive int

	// Close connections after remaining idle for this duration. If the value
	// is zero, then idle connections are not closed. Applications should set
	// the timeout to a value less than the server's timeout.
	IdleTimeout time.Duration

	// If Wait is true and the pool is at the MaxActive limit, then Get() waits
	// for a connection to be returned to the pool before returning.
    // 是否在等待获取空闲连接
	Wait bool

	// Close connections older than this duration. If the value is zero, then
	// the pool does not close connections based on age.
	MaxConnLifetime time.Duration

	chInitialized uint32 // set to 1 when field ch is initialized

	mu     sync.Mutex    // mu protects the following fields
	closed bool          // set to true when the pool is closed.
	active int           // the number of open connections in the pool
	ch     chan struct{} // limits open connections when p.Wait is true
	idle   idleList      // idle connections
}
```

