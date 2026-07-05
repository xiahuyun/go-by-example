package main

import (
	"fmt"
	"sync"
)

func main() {
	var onceA, onceB sync.Once
	var initB func()
	initA := func() {
		fmt.Println("initA")
		onceB.Do(initB)
	}
	initB = func() {
		fmt.Println("initB")
		onceA.Do(initA)
	} // 1
	onceA.Do(initA) // 2
}
