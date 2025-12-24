package main

import (
	"fmt"
	"sync"
	"time"
)

var sharedResource = false

func waiter(cond *sync.Cond, id int) {
	cond.L.Lock()
	for !sharedResource {
		fmt.Printf("Waiter %d is waiting...\n", id)
		cond.Wait()
	}
	fmt.Printf("Waiter %d is proceeding...\n", id)
	cond.L.Unlock()
}

func notifier(cond *sync.Cond) {
	//time.Sleep(2 * time.Second)
	cond.L.Lock()
	sharedResource = true
	fmt.Println("Notifier is notifying all waiters...")
	//cond.Broadcast()
	cond.Signal()
	cond.L.Unlock()
}

func main() {
	var mutex sync.Mutex
	var cond = sync.NewCond(&mutex)

	for i := 1; i <= 3; i++ {
		go waiter(cond, i)
	}

	go notifier(cond)

	time.Sleep(5 * time.Second)
	fmt.Println("Main goroutine is exiting.")
}
