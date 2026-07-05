package main

import (
	"sync"
)

type value struct {
	mu    sync.Mutex
	value int
}

func main() {
	var wg sync.WaitGroup

	/*
		printSum := func(v1, v2 *value) {
			defer wg.Done()
			v1.mu.Lock()         //1
			defer v1.mu.Unlock() //2

			time.Sleep(1 * time.Second) //3
			v2.mu.Lock()
			defer v2.mu.Unlock()

			fmt.Printf("sum=%v\n", v1.value+v2.value)
		}

		var a, b value
		wg.Add(2)
		go printSum(&a, &b)
		go printSum(&a, &b)
		wg.Wait()
	*/

	// dead-lock
	aSum := func(a, b *value) {
		b.mu.Lock()
		a.mu.Lock()
		a.mu.Unlock()
		b.mu.Unlock()

	}
	bSum := func(a, b *value) {
		a.mu.Lock()
		b.mu.Lock()
		b.mu.Unlock()
		a.mu.Unlock()
	}

	var c, d value
	wg.Add(2)
	go aSum(&c, &d)
	go bSum(&c, &d)
	wg.Wait()

}
