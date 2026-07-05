package main

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

func main() {
	filePath := "data.db"

	// 1 打开文件
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	size := 4096

	fileInfo, err := file.Stat()
	if err != nil {
		panic(err)
	}
	fileSize := fileInfo.Size()
	fmt.Printf("file size: %d\n", fileSize)

	// 2 扩展文件大小
	err = file.Truncate(int64(size))
	if err != nil {
		panic(err)
	}

	// 3 mmap 文件
	data, err := unix.Mmap(
		int(file.Fd()),
		0,
		size,
		unix.PROT_READ|unix.PROT_WRITE,
		unix.MAP_SHARED,
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("mmap success")

	// 4 写入数据
	copy(data[0:], []byte("hello mmap"))

	// 5 读取数据
	fmt.Println(string(data))

	// 6 强制刷盘
	err = unix.Msync(data, unix.MS_SYNC)
	if err != nil {
		panic(err)
	}

	// 7 解除映射
	err = unix.Munmap(data)
	if err != nil {
		panic(err)
	}

	os.Remove(filePath)
	fmt.Println("done")
}
