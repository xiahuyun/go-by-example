package main

import (
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"
)

// 文件锁结构
type FileLocker struct {
	file *os.File
}

// 创建新的文件锁
func NewFileLocker(filename string) (*FileLocker, error) {
	// 以读写模式打开或创建文件
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	return &FileLocker{file: file}, nil
}

// 加锁（阻塞）
func (fl *FileLocker) Lock() error {
	return syscall.Flock(int(fl.file.Fd()), syscall.LOCK_EX)
}

// 尝试加锁（非阻塞）
func (fl *FileLocker) TryLock() error {
	return syscall.Flock(int(fl.file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
}

// 解锁
func (fl *FileLocker) Unlock() error {
	return syscall.Flock(int(fl.file.Fd()), syscall.LOCK_UN)
}

// 关闭文件
func (fl *FileLocker) Close() error {
	return fl.file.Close()
}

func main() {
	fmt.Println("=== 文件锁演示示例 ===")

	var wg sync.WaitGroup

	// 模拟多个协程竞争文件锁
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// 为每个协程创建独立的锁实例
			localLocker, err := NewFileLocker("test.lock")
			if err != nil {
				fmt.Printf("协程%d: 创建锁失败: %v\n", id, err)
				return
			}
			defer localLocker.Close()

			fmt.Printf("协程%d: 尝试获取锁...\n", id)

			// 第一个协程先获取锁并持有2秒
			if id == 1 {
				fmt.Printf("协程%d: 获取排他锁...\n", id)
				if err := localLocker.Lock(); err != nil {
					fmt.Printf("协程%d: 加锁失败: %v\n", id, err)
					return
				}
				fmt.Printf("协程%d: 已获得锁，开始处理数据...\n", id)
				time.Sleep(5 * time.Second) // 模拟耗时操作
				fmt.Printf("协程%d: 处理完成，释放锁\n", id)
				localLocker.Unlock()
				return
			}

			// 其他协程等待0.5秒后尝试
			time.Sleep(500 * time.Millisecond)

			// 尝试非阻塞加锁
			fmt.Printf("协程%d: 尝试非阻塞加锁...\n", id)
			if err := localLocker.TryLock(); err != nil {
				if err == syscall.EWOULDBLOCK {
					fmt.Printf("协程%d: 锁被占用，等待后重试...\n", id)
					// 等待后使用阻塞锁
					time.Sleep(1 * time.Second)
					fmt.Printf("协程%d: 重新尝试加锁...\n", id)
					if err := localLocker.Lock(); err != nil {
						fmt.Printf("协程%d: 加锁失败: %v\n", id, err)
						return
					}
					fmt.Printf("协程%d: 已获得锁\n", id)
				} else {
					fmt.Printf("协程%d: 加锁失败: %v\n", id, err)
					return
				}
			} else {
				fmt.Printf("协程%d: 已获得锁\n", id)
			}

			// 模拟操作
			time.Sleep(500 * time.Millisecond)
			fmt.Printf("协程%d: 操作完成，释放锁\n", id)
			localLocker.Unlock()
		}(i)
	}

	wg.Wait()
	fmt.Println("=== 演示结束 ===")

	// 清理
	os.Remove("test.lock")
}
