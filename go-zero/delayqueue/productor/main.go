package main

import (
	"fmt"
	"time"

	"github.com/zeromicro/go-queue/dq"
)

func main() {
	producer := dq.NewProducer([]dq.Beanstalk{
		{
			Endpoint: "localhost:11300",
			Tube:     "tube",
		},
		{
			Endpoint: "localhost:11301",
			Tube:     "tube",
		},
	})

	// 延迟 5s 后处理
	_, err := producer.Delay([]byte("hello"), time.Second*5)
	if err != nil {
		fmt.Println(err)
	}

	// 在指定时间点处理
	_, err = producer.At([]byte("world"), time.Now().Add(time.Second*10))
	if err != nil {
		fmt.Println(err)
	}
}
