package main

import (
	configcenter "github.com/zeromicro/go-zero/core/configcenter"
	"github.com/zeromicro/go-zero/core/configcenter/subscriber"
)

// 配置结构定义
type TestSt struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	// 创建 etcd subscriber
	ss := subscriber.MustNewEtcdSubscriber(subscriber.EtcdConf{
		Hosts: []string{"localhost:2379"}, // etcd 地址
		Key:   "test1",                    // 配置key
	})

	// 创建 configcenter
	// 修改为text类型，因为从etcd获取的值'xiahuyun'不是JSON格式
	cc := configcenter.MustNewConfigCenter[TestSt](configcenter.Config{
		Type: "json", // 配置值类型：json,yaml,toml,text
	}, ss)

	// 获取配置
	// 注意: 配置如果发生变更，调用的结果永远获取到最新的配置
	v, err := cc.GetConfig()
	if err != nil {
		panic(err)
	}
	println(v.Name, v.Age)

	// 如果想监听配置变化，可以添加 listener
	cc.AddListener(func() {
		v, err := cc.GetConfig()
		if err != nil {
			panic(err)
		}
		println(v.Name, v.Age)
	})

	select {}
}
