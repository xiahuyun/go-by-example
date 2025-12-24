package main

import (
	"sync"
	"testing"
)

var globalBuf []byte

func BenchmarkAllocateWithoutPool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buf := make([]byte, 1024)
		globalBuf = buf
	}
}

func BenchmarkAllocateWithPool(b *testing.B) {
	pool := sync.Pool{New: func() interface{} { return make([]byte, 1024) }}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := pool.Get().([]byte)
		pool.Put(buf)
	}
}
