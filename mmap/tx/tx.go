package main

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"
)

// 用 mmap 映射文件
func mmapFile(f *os.File) ([]byte, error) {
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if fi.Size() == 0 {
		return nil, errors.New("empty file")
	}
	return syscall.Mmap(int(f.Fd()), 0, int(fi.Size()),
		syscall.PROT_READ, syscall.MAP_SHARED)
}

func main() {
	// 1. 创建测试数据库文件
	f, _ := os.Create("test.db")
	f.WriteString("PAGE1:OLD-DATA    PAGE2:OLD-DATA") // 模拟两个旧页
	f.Close()

	// 2. 重新打开 + mmap（模拟 旧读事务 长期持有 mmap）
	roFile, _ := os.Open("test.db")
	defer roFile.Close()
	mmapData, _ := mmapFile(roFile)
	fmt.Println("【旧读事务 mmap 初始】:", string(mmapData))

	// 3. 启动后台读事务，一直读 mmap 数据
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 5; i++ {
			fmt.Printf("[旧读事务 mmap 读取]: %q\n", string(mmapData))
			time.Sleep(1 * time.Second)
		}
	}()

	time.Sleep(1 * time.Second)
	fmt.Println("\n===== 写事务覆写磁盘物理页 =====")

	// 4. 写事务：直接覆写磁盘文件（模拟：合并小页 → 释放 → 复用）
	wf, _ := os.OpenFile("test.db", os.O_WRONLY|os.O_TRUNC, 0644)
	wf.WriteString("PAGE1:NEW-DATA    PAGE2:NEW-DATA")
	wf.Close()

	// 5. 等待读事务看到脏数据
	wg.Wait()

	// 清理
	syscall.Munmap(mmapData)
	os.Remove("test.db")
	fmt.Println("\n✅ 演示完成：mmap 被磁盘覆写污染！")
}
