package main

import "time"

func main() {
	var count int
	go func() {
		count++
	}()
	go func() {
		count++
	}()

	time.Sleep(1 * time.Second)
}
