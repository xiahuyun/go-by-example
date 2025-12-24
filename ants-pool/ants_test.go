package main

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/panjf2000/ants/v2"
)

const (
	n          = 10000000
	BenchParam = 100
)

const (
	_   = 1 << (10 * iota)
	KiB // 1024
	MiB // 1048576
)

func demoFunc() {
	time.Sleep(time.Duration(BenchParam) * time.Millisecond)
}

func TestNoPool(t *testing.T) {
	var start, end runtime.MemStats
	runtime.ReadMemStats(&start)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			demoFunc()
			wg.Done()
		}()
	}
	wg.Wait()
	runtime.ReadMemStats(&end)
	usageMem := end.TotalAlloc/MiB - start.TotalAlloc/MiB
	t.Logf("memory usage:%d MB", usageMem)
}

func TestWithAntsPool(t *testing.T) {
	var start, end runtime.MemStats
	runtime.ReadMemStats(&start)

	p, _ := ants.NewPool(10000, ants.WithNonblocking(false))
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		_ = p.Submit(func() {
			demoFunc()
			wg.Done()
		})
	}
	wg.Wait()
	p.Release()

	runtime.ReadMemStats(&end)
	usageMem := end.TotalAlloc/MiB - start.TotalAlloc/MiB
	t.Logf("memory usage:%d MB", usageMem)
}

func TestNoPoolThroughput(t *testing.T) {
	var start, end runtime.MemStats
	runtime.ReadMemStats(&start)
	for i := 0; i < n; i++ {
		go func() {
			demoFunc()
		}()
	}
	runtime.ReadMemStats(&end)
	usageMem := end.TotalAlloc/MiB - start.TotalAlloc/MiB
	t.Logf("memory usage:%d MB", usageMem)
}

func TestWithAntsPoolThroughput(t *testing.T) {
	var start, end runtime.MemStats
	runtime.ReadMemStats(&start)

	p, _ := ants.NewPool(10000, ants.WithNonblocking(false))
	for i := 0; i < n; i++ {
		_ = p.Submit(func() {
			demoFunc()
		})
	}
	p.Release()

	runtime.ReadMemStats(&end)
	usageMem := end.TotalAlloc/MiB - start.TotalAlloc/MiB
	t.Logf("memory usage:%d MB", usageMem)
}

func demoFunc2() {
	for i := 0; i < 10000; i++ {
		i++
	}
}

func TestNoPoolWithFunc(t *testing.T) {
	var start, end runtime.MemStats
	runtime.ReadMemStats(&start)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			demoFunc2()
			wg.Done()
		}()
	}
	wg.Wait()
	runtime.ReadMemStats(&end)
	usageMem := end.TotalAlloc/MiB - start.TotalAlloc/MiB
	t.Logf("memory usage:%d MB", usageMem)
}

func TestWithAntsPoolWithFunc(t *testing.T) {
	var start, end runtime.MemStats
	runtime.ReadMemStats(&start)

	p, _ := ants.NewPool(10000, ants.WithNonblocking(false))
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		_ = p.Submit(func() {
			demoFunc2()
			wg.Done()
		})
	}
	wg.Wait()
	p.Release()

	runtime.ReadMemStats(&end)
	usageMem := end.TotalAlloc/MiB - start.TotalAlloc/MiB
	t.Logf("memory usage:%d MB", usageMem)
}
