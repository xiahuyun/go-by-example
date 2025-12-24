package main

import (
	"fmt"

	"github.com/zeromicro/go-zero/core/mr"
)

func main() {
	generateFunc := func(source chan<- int) {
		for i := 0; i < 10; i++ {
			source <- i
		}
	}

	mapperFunc := func(item int, writer mr.Writer[int], cancel func(error)) {
		writer.Write(item * 2)
	}

	reducerFunc := func(pipe <-chan int, writer mr.Writer[int], cancel func(error)) {
		sum := 0
		for v := range pipe {
			sum += v
		}
		writer.Write(sum)
	}

	result, err := mr.MapReduce(generateFunc, mapperFunc, reducerFunc, mr.WithWorkers(4))
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Result:", result) // Output: Result: 90
	}

}
