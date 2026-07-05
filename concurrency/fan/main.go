package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

// isPrime 判断一个整数是否为素数
func isPrime(n int) bool {
	/*
		if n < 2 {
			return false
		}
		if n == 2 {
			return true
		}
		if n%2 == 0 {
			return false
		}
		// 检查奇数因子
		for i := 3; i*i <= n; i += 2 {
			if n%i == 0 {
				return false
			}
		}
	*/

	for i := 2; i < n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

// primeFinder 从整数流中过滤出素数并发送到输出流
func primeFinder(done <-chan interface{}, intStream <-chan int) <-chan int {
	primeStream := make(chan int)
	go func() {
		defer close(primeStream)
		for {
			select {
			case <-done:
				return
			case num, ok := <-intStream:
				if !ok {
					return // 输入流已关闭
				}
				if isPrime(num) {
					select {
					case <-done:
						return
					case primeStream <- num:
					}
				}
			}
		}
	}()
	return primeStream
}

func main() {
	repeatFn := func(done <-chan interface{}, fn func() interface{}) <-chan interface{} {
		valueStream := make(chan interface{})
		go func() {
			defer close(valueStream)
			for {
				select {
				case <-done:
					return
				case valueStream <- fn():
				}
			}
		}()
		return valueStream
	}

	take := func(done <-chan interface{}, valueStream <-chan int, num int) <-chan int {
		takeStream := make(chan int)
		go func() {
			defer close(takeStream)
			for i := 0; i < num; i++ {
				select {
				case <-done:
					return
				case takeStream <- <-valueStream:
				}
			}
		}()
		return takeStream
	}

	toInt := func(done <-chan interface{}, valueStream <-chan interface{}) <-chan int {
		intStream := make(chan int)
		go func() {
			defer close(intStream)
			for v := range valueStream {
				select {
				case <-done:
					return
				case intStream <- v.(int):
				}
			}
		}()
		return intStream
	}

	fanIn := func(done <-chan interface{}, channels ...<-chan int) <-chan int {

		var wg sync.WaitGroup
		multiplexedStream := make(chan int)

		multiplex := func(c <-chan int) {
			defer wg.Done()
			for i := range c {
				select {
				case <-done:
					return
				case multiplexedStream <- i:
				}
			}
		}

		// 从所有的通道中取数据
		wg.Add(len(channels))
		for _, c := range channels {
			go multiplex(c)
		}

		// 等待所有数据汇总完毕
		go func() {
			wg.Wait()
			close(multiplexedStream)
		}()

		return multiplexedStream
	}

	rand := func() interface{} { return rand.Intn(50000000) }

	done := make(chan interface{})
	defer close(done)

	start := time.Now()

	randIntStream := toInt(done, repeatFn(done, rand))
	fmt.Println("Primes:")

	numFinders := runtime.NumCPU()
	finders := make([]<-chan int, numFinders)
	for i := 0; i < numFinders; i++ {
		finders[i] = primeFinder(done, randIntStream)
	}

	for prime := range take(done, fanIn(done, finders...), 10) {
		fmt.Printf("\t%d\n", prime)
	}

	/*
		for prime := range take(done, primeFinder(done, randIntStream), 20) {
			fmt.Printf("\t%d\n", prime)
		}
	*/

	fmt.Printf("Search took: %v", time.Since(start))
}
