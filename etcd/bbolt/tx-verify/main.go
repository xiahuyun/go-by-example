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
	dbFile = "tx-verify.db"
	bucket = []byte("bench")
)

type scenarioResult struct {
	name          string
	waitToEnterTx time.Duration
	callbackCost  time.Duration
	waitAfterCb   time.Duration
	totalCallCost time.Duration
	dbSizeBefore  int64
	dbSizeAfter   int64
	dbSizeDelta   int64
	judgement     string
}

func main() {
	if err := os.Remove(dbFile); err != nil && !os.IsNotExist(err) {
		log.Fatalf("remove old db failed: %v", err)
	}

	db, err := bolt.Open(dbFile, 0o600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatalf("open db failed: %v", err)
	}
	defer db.Close()

	if err := db.Update(func(tx *bolt.Tx) error {
		_, e := tx.CreateBucketIfNotExists(bucket)
		return e
	}); err != nil {
		log.Fatalf("init bucket failed: %v", err)
	}

	hold := 2 * time.Second

	fmt.Println("=== 实验 1: 长读事务进行中，写事务会不会等待读事务释放？ ===")
	r1, err := readBlocksWrite(db, hold)
	if err != nil {
		log.Fatalf("scenario 1 failed: %v", err)
	}
	printResult(r1)

	fmt.Println("\n=== 实验 2: 长写事务进行中，读事务会不会等待写事务结束？ ===")
	r2, err := writeBlocksRead(db, hold)
	if err != nil {
		log.Fatalf("scenario 2 failed: %v", err)
	}
	printResult(r2)
}

func readBlocksWrite(db *bolt.DB, hold time.Duration) (scenarioResult, error) {
	readerReady := make(chan struct{})
	readerDone := make(chan error, 1)

	go func() {
		err := db.View(func(tx *bolt.Tx) error {
			_ = tx.Bucket(bucket)
			fmt.Printf("[reader] 已进入只读事务，持有 %v\n", hold)
			close(readerReady)
			time.Sleep(hold)
			return nil
		})
		readerDone <- err
	}()

	<-readerReady
	sizeBefore, err := dbFileSize(dbFile)
	if err != nil {
		return scenarioResult{}, err
	}

	start := time.Now()
	writeEntered := make(chan time.Time, 1)
	writeLeft := make(chan time.Time, 1)

	err = db.Update(func(tx *bolt.Tx) error {
		writeEntered <- time.Now()
		b := tx.Bucket(bucket)
		// 写入较大 value，尽量触发文件增长和提交阶段的真实开销。
		bigVal := bytes.Repeat([]byte("x"), 32<<20) // 32 MiB
		if err := b.Put([]byte("big-k1"), bigVal); err != nil {
			return err
		}
		writeLeft <- time.Now()
		return nil
	})
	if err != nil {
		return scenarioResult{}, err
	}

	enterAt := <-writeEntered
	leftAt := <-writeLeft
	wait := enterAt.Sub(start)
	cbCost := leftAt.Sub(enterAt)
	total := time.Since(start)
	postCbWait := total - (wait + cbCost)
	sizeAfter, err := dbFileSize(dbFile)
	if err != nil {
		return scenarioResult{}, err
	}
	sizeDelta := sizeAfter - sizeBefore

	if err := <-readerDone; err != nil {
		return scenarioResult{}, err
	}

	judge := "结论: 写事务在提交阶段明显等待了读事务（存在阻塞）"
	if wait < hold/2 && postCbWait < hold/2 {
		judge = "结论: 写事务基本未被读事务阻塞"
	}

	return scenarioResult{
		name:          "读 -> 写",
		waitToEnterTx: wait,
		callbackCost:  cbCost,
		waitAfterCb:   postCbWait,
		totalCallCost: total,
		dbSizeBefore:  sizeBefore,
		dbSizeAfter:   sizeAfter,
		dbSizeDelta:   sizeDelta,
		judgement:     judge,
	}, nil
}

func writeBlocksRead(db *bolt.DB, hold time.Duration) (scenarioResult, error) {
	writerReady := make(chan struct{})
	writerDone := make(chan error, 1)

	go func() {
		err := db.Update(func(tx *bolt.Tx) error {
			fmt.Printf("[writer] 已进入写事务，持有 %v\n", hold)
			close(writerReady)
			time.Sleep(hold)
			b := tx.Bucket(bucket)
			return b.Put([]byte("k2"), []byte("v2"))
		})
		writerDone <- err
	}()

	<-writerReady
	sizeBefore, err := dbFileSize(dbFile)
	if err != nil {
		return scenarioResult{}, err
	}

	start := time.Now()
	readEntered := make(chan time.Time, 1)
	readLeft := make(chan time.Time, 1)

	err = db.View(func(tx *bolt.Tx) error {
		readEntered <- time.Now()
		b := tx.Bucket(bucket)
		_ = b.Get([]byte("k2"))
		readLeft <- time.Now()
		return nil
	})
	if err != nil {
		return scenarioResult{}, err
	}

	enterAt := <-readEntered
	leftAt := <-readLeft
	wait := enterAt.Sub(start)
	cbCost := leftAt.Sub(enterAt)
	total := time.Since(start)

	if err := <-writerDone; err != nil {
		return scenarioResult{}, err
	}
	sizeAfter, err := dbFileSize(dbFile)
	if err != nil {
		return scenarioResult{}, err
	}
	sizeDelta := sizeAfter - sizeBefore

	judge := "结论: 读事务被写事务明显阻塞"
	if wait < hold/2 && total < hold/2 {
		judge = "结论: 读事务基本未被写事务阻塞（可并发读取旧快照）"
	}

	return scenarioResult{
		name:          "写 -> 读",
		waitToEnterTx: wait,
		callbackCost:  cbCost,
		totalCallCost: total,
		dbSizeBefore:  sizeBefore,
		dbSizeAfter:   sizeAfter,
		dbSizeDelta:   sizeDelta,
		judgement:     judge,
	}, nil
}

func printResult(r scenarioResult) {
	fmt.Printf("场景: %s\n", r.name)
	fmt.Printf("等待进入事务耗时: %v\n", r.waitToEnterTx)
	if r.callbackCost > 0 {
		fmt.Printf("事务回调执行耗时: %v\n", r.callbackCost)
	}
	if r.waitAfterCb > 0 {
		fmt.Printf("回调结束后等待耗时: %v\n", r.waitAfterCb)
	}
	fmt.Printf("整个调用总耗时: %v\n", r.totalCallCost)
	fmt.Printf("DB 文件大小变化: %d -> %d (delta=%d)\n", r.dbSizeBefore, r.dbSizeAfter, r.dbSizeDelta)
	if r.dbSizeDelta > 0 {
		fmt.Println("文件发生增长：本次写路径大概率触发了 grow/mmap 相关流程。")
	} else {
		fmt.Println("文件未增长：本次写路径未观察到 grow/mmap 导致的文件扩容。")
	}
	fmt.Println(r.judgement)
}

func dbFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}
