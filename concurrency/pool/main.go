package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	var numCalcsCreated int32
	calcPool := &sync.Pool{
		New: func() interface{} {
			atomic.AddInt32(&numCalcsCreated, 1)
			mem := make([]byte, 1024)
			return &mem // 1
		},
	}

	// 将池扩充到4KB
	calcPool.Put(calcPool.New())
	calcPool.Put(calcPool.New())
	calcPool.Put(calcPool.New())
	calcPool.Put(calcPool.New())
	const numWorkers = 1024 * 1024
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := numWorkers; i > 0; i-- {
		go func() {
			defer wg.Done()

			mem := calcPool.Get().(*[]byte) // 2
			defer calcPool.Put(mem)

		}()
	}

	// 假设内存中执行了一些快速的操作

	wg.Wait()
	fmt.Printf("%d calculators were created.", atomic.LoadInt32(&numCalcsCreated))
}
