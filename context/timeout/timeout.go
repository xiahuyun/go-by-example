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
			println(fmt.Sprintf("task %d is canceled", id))
			return
		case <-ticker.C:
			println(fmt.Sprintf("running task %d", id))
		}
	}
}

func main() {
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)

	go runningTask(ctx, 1)
	go runningTask(ctx, 2)

	time.Sleep(2 * time.Second)
	fmt.Println("main goroutine is done")

	//cancel()
	time.Sleep(1 * time.Second)
}
