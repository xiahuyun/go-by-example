package main

import (
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"
)

func task(id int) error {
	time.Sleep(500 * time.Millisecond)
	if id == 2 { // 模拟任务2失败
		return fmt.Errorf("task %d failed", id)
	}
	if id == 4 { // 模拟任务4失败
		return fmt.Errorf("task %d failed", id)
	}
	fmt.Printf("Task %d succeeded\n", id)
	return nil
}

func main() {
	var g errgroup.Group
	for i := 1; i <= 5; i++ {
		id := i
		g.Go(func() error {
			return task(id)
		})
	}

	if err := g.Wait(); err != nil {
		fmt.Printf("Error occurred: %v\n", err)
	} else {
		fmt.Println("All tasks completed successfully")
	}
}
