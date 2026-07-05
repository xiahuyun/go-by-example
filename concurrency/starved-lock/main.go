package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

var wg sync.WaitGroup
var sharedLock sync.Mutex

const runtime1 = 1 * time.Second

func main() {
	runtime.GOMAXPROCS(1) //1

	greedyWorker := func() {
		defer wg.Done()

		//var count int

		sharedLock.Lock()
		fmt.Println("xhy got the lock")
		for {
		}

		// fmt.Printf("Greedy worker was able to execute %v work loops\n", count)
	}

	politeWorker := func() {
		defer wg.Done()

		time.Sleep(10 * time.Millisecond) //1

		var count int

		for begin := time.Now(); time.Since(begin) <= runtime1; {
			sharedLock.Lock()
			fmt.Println("yx got the lock")
			time.Sleep(1 * time.Nanosecond)
			sharedLock.Unlock()

			sharedLock.Lock()
			time.Sleep(1 * time.Nanosecond)
			sharedLock.Unlock()

			sharedLock.Lock()
			time.Sleep(1 * time.Nanosecond)
			sharedLock.Unlock()

			count++
		}
		fmt.Printf("Polite worker was able to execute %v work loops.\n", count)
	}

	wg.Add(2)
	go greedyWorker()
	go politeWorker()

	wg.Wait()
}
