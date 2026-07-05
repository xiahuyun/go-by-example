package main

import "fmt"

func main() {
	doWork := func(strings <-chan string) <-chan interface{} {
		completed := make(chan interface{})
		go func() {
			defer fmt.Println("doWork exited.")
			defer close(completed)
			for s := range strings {
				fmt.Println(s)
			}
		}()
		return completed
	}

	/*
		// deadlock
		<-doWork(nil)
		// 这里还有其他任务执行
		fmt.Println("Done.")
	*/

	go func() {
		<-doWork(nil)
		fmt.Println("Done.")
	}()

	fmt.Println("Other tasks are running...")
	/*
			Other tasks are running...
		fatal error: all goroutines are asleep - deadlock!

		goroutine 1 [select (no cases)]:
		main.main()
		        /Users/hxia/project/go-by-example/concurrency/leak/main.go:31 +0x74

		goroutine 5 [chan receive]:
		main.main.func2()
		        /Users/hxia/project/go-by-example/concurrency/leak/main.go:26 +0x30
		created by main.main in goroutine 1
		        /Users/hxia/project/go-by-example/concurrency/leak/main.go:25 +0x3c

		goroutine 33 [chan receive (nil chan)]:
		main.main.func1.1()
		        /Users/hxia/project/go-by-example/concurrency/leak/main.go:11 +0xcc
		created by main.main.func1 in goroutine 5
		        /Users/hxia/project/go-by-example/concurrency/leak/main.go:8 +0x78
		exit status 2

				select {}
	*/

	for {
	}
}
