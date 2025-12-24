package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
)

func main() {
	// Create a goroutine pool with size 10
	pool, _ := ants.NewPool(10, ants.WithNonblocking(true))
	defer pool.Release()

	var wg sync.WaitGroup

	// Submit 20 tasks
	for i := 0; i < 20; i++ {
		wg.Add(1)
		task := func(i int) func() {
			return func() {
				defer wg.Done()
				fmt.Printf("Task %d is running\n", i)
				time.Sleep(10 * time.Second)
				fmt.Printf("Task %d completed\n", i)
			}
		}(i)

		if err := pool.Submit(task); err != nil {
			wg.Done()
		}
	}

	// Wait for all tasks to complete
	fmt.Println("Waiting for tasks to finish...")
	wg.Wait()
	fmt.Println("All tasks completed")
}
