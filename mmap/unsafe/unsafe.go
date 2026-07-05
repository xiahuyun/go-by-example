package main

import (
	"fmt"
	"unsafe"
)

func main() {
	// 模拟 mmap 得到的一块内存
	b := make([]byte, 16)

	// 初始化数据
	for i := 0; i < len(b); i++ {
		b[i] = byte(i + 1)
	}

	fmt.Println("原始 slice:", b)

	// 🔥 核心转换（类似 bbolt）
	data := (*[1024]byte)(unsafe.Pointer(&b[0]))

	// 通过“数组指针”访问
	fmt.Println("data[0] =", data[0])
	fmt.Println("data[5] =", data[5])

	// 修改数据（通过 data）
	data[0] = 100
	data[1] = 101

	fmt.Println("修改后 slice:", b)

	// 🚨 越界访问（危险！这里只是演示）
	fmt.Println("访问超出 slice 长度但未崩:", data[20])
}
