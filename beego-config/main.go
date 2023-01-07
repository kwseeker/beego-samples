package main

import (
	"github.com/beego/beego/v2/core/config"
	_ "github.com/beego/beego/v2/core/config/etcd"
	_ "github.com/beego/beego/v2/core/config/yaml"
	"github.com/beego/beego/v2/core/logs"
)

func main() {
	// 1 使用默认的配置方式
	a, _ := config.String("author")
	logs.Info("load config: author -> ", a)
	p, _ := config.Int("httpport")
	logs.Info("load config: httpport -> ", p)
	date, _ := config.String("date")
	logs.Info("load config: date -> ", date)

	// 2 使用 config.InitGlobalInstance 指定不同的配置文件类型
	err2 := config.InitGlobalInstance("yaml", "./conf/app.yaml")
	if err2 != nil {
		logs.Critical("An error occurred:", err2)
		panic(err2)
	}

	a2, _ := config.String("author")
	logs.Info("load config: author -> ", a2)
	p2, _ := config.Int("httpport")
	logs.Info("load config: httpport -> ", p2)

	// 3 使用 config.InitGlobalInstance 指定不同的配置类型, 注意时间类的配置单位是纳秒
	err3 := config.InitGlobalInstance("etcd", `{
	  "endpoints": [
		"127.0.0.1:2379"
	  ],
	  "dial-timeout": 5000000
	}`)
	if err3 != nil {
		logs.Critical("An error occurred:", err3)
		panic(err3)
	}

	a3, _ := config.String("author")
	logs.Info("load config: author -> ", a3)

	// 4 使用 Configer 实例
	cfg, err4 := config.NewConfig("ini", "./conf/ext.ini")
	if err4 != nil {
		logs.Error(err4)
	}

	a4, _ := cfg.String("author")
	logs.Info("load config: author -> ", a4)

	//后面加载的全局配置会覆盖前面的
	author, _ := config.String("author")
	logs.Info("load global config: author -> ", author)
	date2, _ := config.String("date")
	logs.Info("load global config: date -> ", date2)
}
