package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"time"

	bolt "go.etcd.io/bbolt"
)

var (
	dbPath     = "tx-no-mmap.db"
	bucketName = []byte("bench")
	keyName    = []byte("hot-key")
)

func main() {
	if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
		log.Fatalf("remove old db failed: %v", err)
	}

	db, err := bolt.Open(dbPath, 0o600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatalf("open db failed: %v", err)
	}
	defer db.Close()

	if err := prepareReusablePages(db); err != nil {
		log.Fatalf("prepare reusable pages failed: %v", err)
	}

	beforeSize, err := fileSize(dbPath)
	if err != nil {
		log.Fatalf("stat before failed: %v", err)
	}

	hold := 2 * time.Second
	readerReady := make(chan struct{})
	readerDone := make(chan error, 1)

	go func() {
		err := db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket(bucketName)
			_ = b.Get(keyName)
			fmt.Printf("[reader] 已进入只读事务，持有 %v\n", hold)
			close(readerReady)
			time.Sleep(hold)
			return nil
		})
		readerDone <- err
	}()

	<-readerReady
	start := time.Now()
	writeEntered := make(chan time.Time, 1)
	writeLeft := make(chan time.Time, 1)

	err = db.Update(func(tx *bolt.Tx) error {
		writeEntered <- time.Now()
		b := tx.Bucket(bucketName)
		// 固定长度更新，目标是尽量复用已有页，避免触发 grow/mmap 重映射。
		if err := b.Put(keyName, bytes.Repeat([]byte("W"), 1024)); err != nil {
			return err
		}
		writeLeft <- time.Now()
		return nil
	})
	if err != nil {
		log.Fatalf("write tx failed: %v", err)
	}

	enterAt := <-writeEntered
	leftAt := <-writeLeft
	waitToEnter := enterAt.Sub(start)
	callbackCost := leftAt.Sub(enterAt)
	total := time.Since(start)

	if err := <-readerDone; err != nil {
		log.Fatalf("reader tx failed: %v", err)
	}

	afterSize, err := fileSize(dbPath)
	if err != nil {
		log.Fatalf("stat after failed: %v", err)
	}

	fmt.Println("=== 结果 ===")
	fmt.Printf("写事务等待进入回调: %v\n", waitToEnter)
	fmt.Printf("写事务回调执行耗时: %v\n", callbackCost)
	fmt.Printf("写事务总耗时: %v\n", total)
	fmt.Printf("DB 文件大小变化: %d -> %d (delta=%d)\n", beforeSize, afterSize, afterSize-beforeSize)

	if total < hold/2 && afterSize == beforeSize {
		fmt.Println("结论: 本次写入未观察到由读事务导致的明显阻塞，且未出现文件扩容。")
		fmt.Println("说明此场景下写入未触发会与读事务强冲突的 mmap 重映射路径。")
	} else {
		fmt.Println("结论: 本次运行出现了明显等待或文件增长，请重跑或降低写入量观察无重映射场景。")
	}
}

func prepareReusablePages(db *bolt.DB) error {
	if err := db.Update(func(tx *bolt.Tx) error {
		_, e := tx.CreateBucketIfNotExists(bucketName)
		return e
	}); err != nil {
		return err
	}

	seedVal := bytes.Repeat([]byte("A"), 1024)
	if err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		for i := 0; i < 4000; i++ {
			k := []byte(fmt.Sprintf("k-%05d", i))
			if err := b.Put(k, seedVal); err != nil {
				return err
			}
		}
		if err := b.Put(keyName, bytes.Repeat([]byte("H"), 1024)); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		for i := 0; i < 3500; i++ {
			k := []byte(fmt.Sprintf("k-%05d", i))
			if err := b.Delete(k); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	// 触发 ReleasePendingPages，使删除产生的 pending pages 转为可复用空闲页。
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		return b.Put([]byte("warmup"), []byte("ok"))
	})
}

func fileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}
