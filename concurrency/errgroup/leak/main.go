package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof" // 注册 pprof 端点
	"time"
)

func leakyWorker(ctx context.Context) {
	interval := 100 * time.Minute
	timer := time.NewTimer(interval)
	defer timer.Stop()
	for {
		// 模拟无限循环，没有任何退出机制
		// time.Sleep(100 * time.Minute)
		timer.Reset(interval)
		select {
		case <-ctx.Done():
			fmt.Println("leaky worker is canceled")
			return
		case <-timer.C:
			fmt.Println("leaky worker is running")
		}
	}
}

func main() {
	// 启动 pprof 服务器（端口 6060）
	go func() {
		fmt.Println("pprof listening on :6060")
		http.ListenAndServe(":6060", nil)
	}()

	// 每次请求就泄漏一个 goroutine
	http.HandleFunc("/leak", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		go leakyWorker(ctx)

		// 模拟一段时间后取消（比如 5 秒后）
		time.AfterFunc(5*time.Second, cancel)

		fmt.Fprintf(w, "leaked one goroutine")
	})

	fmt.Println("server started at :9080")
	if err := http.ListenAndServe(":9080", nil); err != nil {
		fmt.Printf("ListenAndServe: %v\n", err)
	}
}
