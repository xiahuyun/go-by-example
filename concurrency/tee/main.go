package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	fileHandler := func(done <-chan interface{}, valueStream <-chan string) interface{} {
		for {
			select {
			case <-done:
				fmt.Println("fileHandler done")
				return nil
			case v, ok := <-valueStream:
				if !ok {
					fmt.Println("fileHandler valueStream closed")
					return nil
				}
				fmt.Println("fileHandler: ", v)
			}
		}
	}

	consoleHandler := func(done <-chan interface{}, valueStream <-chan string) interface{} {
		for {
			select {
			case <-done:
				fmt.Println("consoleHandler done")
				return nil
			case v, ok := <-valueStream:
				if !ok {
					fmt.Println("consoleHandler valueStream closed")
					return nil
				}
				fmt.Println("consoleHandler: ", v)
			}
		}
	}

	var done = make(chan interface{})

	fileStream, consoleStream := make(chan string), make(chan string)
	go fileHandler(done, fileStream)
	go consoleHandler(done, consoleStream)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			/*
				select {
				case fileStream <- fmt.Sprintf("file %d", i):
				case consoleStream <- fmt.Sprintf("console %d", i):
				}
			*/

			fileStream <- fmt.Sprintf("file %d", i)
			consoleStream <- fmt.Sprintf("console %d", i)
		}
	}()
	wg.Wait()

	time.Sleep(time.Second)
	close(done)
	time.Sleep(time.Second)
	fmt.Println("main done")
	time.Sleep(time.Second)
}
