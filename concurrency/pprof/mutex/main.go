package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"sync"
	"sync/atomic"
)

func main() {
	runtime.SetMutexProfileFraction(1) // 采样所有锁等待事件

	// 启动 pprof 服务器（端口 6060）
	go func() {
		fmt.Println("pprof listening on :6060")
		http.ListenAndServe(":6060", nil)
	}()

	var counter int32
	var wg sync.WaitGroup
	var mu sync.Mutex
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}

	wg.Wait()
	fmt.Printf("Counter: %d\n", atomic.LoadInt32(&counter))
	select {}
}
