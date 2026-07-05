package main

import (
	"fmt"
	"time"
)

func main() {
	ch := make(chan int)

	go func() {
		time.Sleep(2 * time.Second)
		ch <- 1
	}()

	for {
		select {
		case <-ch:
			fmt.Println("ch")
		case <-time.After(3 * time.Second):
			fmt.Println("operation timeout")
			return
		}
	}
}
