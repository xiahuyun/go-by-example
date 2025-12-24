package main

import (
	"fmt"

	"github.com/spaolacci/murmur3"
)

func main() {
	// 定义要哈希的ID
	id := "43153213261643146"

	// 使用 murmur3 算法计算哈希值

	h := murmur3.New64()
	h.Write([]byte(id))
	hashValue := h.Sum64()

	// 取模操作（这里以 1000 为例，可以根据需要修改模数）
	modulus := uint64(5)
	result := hashValue % modulus

	// 输出结果
	fmt.Printf("原始ID: %s\n", id)
	fmt.Printf("哈希值: %d\n", hashValue)
	fmt.Printf("哈希值(十六进制): %x\n", hashValue)
	fmt.Printf("哈希后取模结果 (模数 %d): %d\n", modulus, result)
}
