package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

const (
	fileSize  = 64 * 1024 * 1024 // 64MB
	readCount = 1_000_000        // 100万次读取
	blockSize = 8                // 每次读取8字节
)

func createFile() {
	file, err := os.OpenFile("test.db", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	file.Truncate(fileSize)

	buf := make([]byte, 4096)

	for i := 0; i < fileSize/4096; i++ {
		rand.Read(buf)
		file.Write(buf)
	}
}

func normalRead() {
	file, err := os.Open("test.db")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	buf := make([]byte, blockSize)

	start := time.Now()

	for i := 0; i < readCount; i++ {
		offset := rand.Intn(fileSize - blockSize)
		_, err := file.ReadAt(buf, int64(offset))
		if err != nil {
			panic(err)
		}
	}

	elapsed := time.Since(start)

	fmt.Println("normal read:", elapsed)
}

func mmapRead() {
	file, err := os.Open("test.db")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	data, err := unix.Mmap(
		int(file.Fd()),
		0,
		fileSize,
		unix.PROT_READ,
		unix.MAP_SHARED,
	)
	if err != nil {
		panic(err)
	}

	start := time.Now()

	for i := 0; i < readCount; i++ {
		offset := rand.Intn(fileSize - blockSize)
		_ = data[offset : offset+blockSize]
	}

	elapsed := time.Since(start)

	fmt.Println("mmap read:", elapsed)

	unix.Munmap(data)
}

func main() {

	fmt.Println("creating file...")
	createFile()

	fmt.Println("normal IO test")
	normalRead()

	fmt.Println("mmap test")
	mmapRead()
}
