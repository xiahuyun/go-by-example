package main

import (
	"fmt"
	"runtime"
	"sync"
)

func main() {
	runtime.GOMAXPROCS(1) //1
	var wg sync.WaitGroup
	for _, salutation := range []string{"hello", "greetings", "good day"} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Println(salutation)
		}()
	}
	wg.Wait()
}
