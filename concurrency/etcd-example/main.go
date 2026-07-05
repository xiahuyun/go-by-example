package main

import (
	"fmt"
	"sync"
	"time"
)

// concurrency do things

var cond = sync.NewCond(&sync.Mutex{})
var condition bool

var wait chan struct{} = make(chan struct{})

func eat(id int) {

	cond.L.Lock()

	select {
	case wait <- struct{}{}:
		fmt.Printf("Waiter %d is tell cook to cook fish...\n", id)
	default:
	}

	for !condition {
		fmt.Printf("Waiter %d is waiting...\n", id)
		cond.Wait()
	}

	fmt.Printf("goroutine %d is eating fish\n", id)
	cond.L.Unlock()
}

func cook() {
	for range wait {
		fmt.Println("Cook is cooking fish...")
		// Simulate cooking time
		time.Sleep(2 * time.Second)

		condition = true
		cond.Broadcast()
	}
}

func main() {
	go cook()

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			eat(id)
		}(i)
	}

	wg.Wait()

	fmt.Println("All goroutines are done")
}
