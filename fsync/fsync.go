package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	const filePath = "/tmp/write_fsync_demo.txt"

	// 1. 打开文件（不存在就创建，存在就清空）
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = f.Close()
	}()

	data := []byte(fmt.Sprintf("hello etcd fsync demo at %s\n", time.Now().Format(time.RFC3339Nano)))

	// 2. write：把数据写给操作系统
	// 注意：这里通常只是进入了内核 page cache，不一定已经真正落盘
	n, err := f.Write(data)
	if err != nil {
		panic(err)
	}
	fmt.Printf("write() wrote %d bytes\n", n)

	// 3. 这里可以看到文件内容已经“能读到了”，
	// 但这不等于已经安全落盘
	contentBeforeSync, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	fmt.Printf("content after write(), before fsync(): %q\n", string(contentBeforeSync))

	// 4. fsync：要求操作系统把文件数据真正刷到磁盘
	// 在 Go 里对应 file.Sync()
	start := time.Now()
	if err := f.Sync(); err != nil {
		panic(err)
	}
	fmt.Printf("fsync() done, cost=%s\n", time.Since(start))

	// 5. fsync 之后，这次写入才算“持久化”更有保障
	contentAfterSync, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	fmt.Printf("content after fsync(): %q\n", string(contentAfterSync))

	fmt.Println("\nSummary:")
	fmt.Println("1) Write() usually writes into kernel page cache first.")
	fmt.Println("2) Sync() asks the OS to flush dirty data to storage.")
	fmt.Println("3) Databases like etcd rely on write + fsync before ack/commit.")
}
