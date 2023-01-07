# beego-samples

beego Demo 样例与源码实现分析。

Samples:

+ **beego-hello** 

  最简beego项目。

  ```go
  import "github.com/beego/beego/v2/server/web"
  
  func main() {
  	web.Run()
  }
  ```

+ **beego-config**

  演示分别用ini、yaml配置文件 以及 etcd 作为配置中心的配置方法和原理。

  app.conf 默认用ini格式。

  + 全局配置

    + 使用默认的app.conf

      如果使用默认的app.conf配置，在初始化方法中会默认扫描加载。

      本质是也是借助函数 InitGlobalInstance()：

      ```golang
      Register("ini", &IniConfig{})
      err := InitGlobalInstance("ini", "conf/app.conf")
      ```

    + config.InitGlobalInstance() 指定不同的配置类型

    > 注意不能混用多种类型配置设置全局配置，看源码可以看到存放全局配置的对象 globalInstance，每经过一次配置解析 globalInstance 都会被覆盖，而不是追加配置。
    >
    > ```go
    > func InitGlobalInstance(name string, cfg string) error {
    > 	var err error
    > 	globalInstance, err = NewConfig(name, cfg)
    > 	return err
    > }
    > ```

  + Configer配置实例 - config.NewConfig()

  + 配置中使用环境变量，${环境变量||默认值}

  + 使用 etcd 配置类型

    官方文档的示例代码是错的，而且没说清楚具体使用方法。正确使用方式：

    ```go
    //引入etcd包
    _ "github.com/beego/beego/v2/core/config/etcd"
    
    //第二个参数是json字符串，用于定义 go.etcd.io/etcd/client/v3 客户端配置
    err3 := config.InitGlobalInstance("etcd", `{
    "endpoints": [
    "127.0.0.1:2379"
    ],
    "dial-timeout": 5000000
    }`)
    ```

    > etcd 安装测试:
    >
    > etcd --listen-client-urls 'http://0.0.0.0:2379' --advertise-client-urls 'http://0.0.0.0:2379'
    >
    > ./etcdctl --endpoints=http://127.0.0.1:2379 put author "Arvin Lee"
    >
    > ./etcdctl --endpoints=http://127.0.0.1:2379 get author

+ **beego-web**

  