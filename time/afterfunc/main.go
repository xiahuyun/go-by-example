package main

import (
	"fmt"
	"time"
)

func main() {
	time.AfterFunc(2*time.Second, func() {
		fmt.Println("2 seconds later")
	})

	time.Sleep(5 * time.Second)
}
