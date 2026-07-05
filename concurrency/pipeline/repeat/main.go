package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func main() {
	/*
		repeat := func(done <-chan interface{}, values ...interface{}) <-chan interface{} {
			valueStream := make(chan interface{})
			go func() {
				defer func() { fmt.Println("repeat done"); close(valueStream) }()
				for {
					for _, v := range values {
						select {
						case <-done:
							return
						case valueStream <- v:
						}
					}
				}
			}()
			return valueStream
		}
	*/

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

	take := func(done <-chan interface{}, valueStream <-chan interface{}, num int) <-chan interface{} {
		takeStream := make(chan interface{})
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

	done := make(chan interface{})
	defer close(done)

	randFn := func() interface{} {
		n, err := rand.Int(rand.Reader, big.NewInt(100))
		if err != nil {
			return int64(0)
		}
		return n.Int64()
	}

	for num := range take(done, repeatFn(done, randFn), 10) {
		fmt.Println(num)
	}
}
