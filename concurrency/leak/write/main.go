package main

import (
	"fmt"
)

func main() {
	newRandStream := func() <-chan int {
		randStream := make(chan int)
		go func() {
			defer fmt.Println("newRandStream closure exited.") // 1
			defer close(randStream)
			for {
				randStream <- 1
			}
		}()

		return randStream
	}

	randStream := newRandStream()
	fmt.Println("3 random ints:")
	for i := 1; i <= 3; i++ {
		fmt.Printf("%d: %d\n", i, <-randStream)
	}

	for {
	}
}
