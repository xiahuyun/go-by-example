package main

import (
	"context"
	"fmt"
	"time"
)

func runningTask(ctx context.Context, id int) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			println(fmt.Sprintf("running task %d", id))
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go runningTask(ctx, 1)
	go runningTask(ctx, 2)

	time.Sleep(2 * time.Second)
	fmt.Println("main goroutine is done")
}
