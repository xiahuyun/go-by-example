package main

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
)

type Worker struct{}

func (w *Worker) Name() string {
	return "worker"
}

func main() {
	runtime.GOMAXPROCS(8)

	var createWorkerTime int32
	workerPool := sync.Pool{New: func() interface{} {
		atomic.AddInt32(&createWorkerTime, 1)
		return Worker{}
	}}

	currencyCount := 1024 * 1
	var wg sync.WaitGroup
	for i := 0; i < currencyCount; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			worker := workerPool.Get().(Worker)
			defer workerPool.Put(worker)
			name := worker.Name()
			fmt.Println("worker name: ", name, "currency id: ", i)
			//time.Sleep(time.Millisecond * 100)
		}(i)
	}

	wg.Wait()
	fmt.Println("create worker time: ", atomic.LoadInt32(&createWorkerTime))

	/*
			for i := 0; i < currencyCount; i++ {
				worker := workerPool.Get().(Worker)
				time.Sleep(time.Millisecond * 1)
				workerPool.Put(worker)
			}


		fmt.Println("create worker time: ", atomic.LoadInt32(&createWorkerTime))
	*/
}
