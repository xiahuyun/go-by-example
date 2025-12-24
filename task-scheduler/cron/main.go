package main

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

func main() {
	// 创建 Cron 实例（v3 版本支持秒级精度，默认使用标准 Cron 语法）
	c := cron.New(
		cron.WithSeconds(),          // 启用秒级精度（可选）
		cron.WithLocation(time.UTC), // 设置时区（可选）
	)

	// 添加任务：每 5 秒执行一次（Cron 表达式：0/5 * * * * *）
	_, err := c.AddFunc("0/5 * * * * *", func() {
		fmt.Println("任务执行时间：", time.Now().Format("2006-01-02 15:04:05"))
	})
	if err != nil {
		panic(err)
	}

	// 启动调度器
	c.Start()

	// 阻塞主协程（防止程序退出）
	select {}
}
