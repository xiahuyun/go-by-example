package main

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

func main() {

	fileSize := 512 * 1024 * 1024 // 512MB

	file, err := os.OpenFile("mmap_test.db", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	err = file.Truncate(int64(fileSize))
	if err != nil {
		panic(err)
	}

	data, err := unix.Mmap(
		int(file.Fd()),
		0,
		fileSize,
		unix.PROT_READ|unix.PROT_WRITE,
		unix.MAP_SHARED,
	)

	if err != nil {
		panic(err)
	}

	fmt.Println("mmap done")

	// 只访问前 4KB
	// data[0] = 1
	for i := 0; i < fileSize; i += 1 {
		data[i] = 1
	}

	fmt.Println("pid:", os.Getpid())

	// 不退出，方便观察
	for {
		time.Sleep(time.Second)
	}
}
