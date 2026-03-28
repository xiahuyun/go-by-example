package treebench

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/google/btree"
	"github.com/petar/GoLLRB/llrb"
)

const (
	btreeDegree = 32
)

var (
	insertSizes = []int{10_000, 100_000}
	querySizes  = []int{10_000, 100_000}
	deleteSizes = []int{10_000, 50_000}
)

type btreeIntItem int

func (a btreeIntItem) Less(b btree.Item) bool {
	return a < b.(btreeIntItem)
}

type llrbIntItem int

func (a llrbIntItem) Less(b llrb.Item) bool {
	return a < b.(llrbIntItem)
}

func BenchmarkInsertRandom(b *testing.B) {
	for _, size := range insertSizes {
		keys := shuffledRange(size, 2026)

		b.Run(fmt.Sprintf("BTree/N=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				tree := btree.New(btreeDegree)
				for _, key := range keys {
					tree.ReplaceOrInsert(btreeIntItem(key))
				}
			}
		})

		b.Run(fmt.Sprintf("LLRB/N=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				tree := llrb.New()
				for _, key := range keys {
					tree.ReplaceOrInsert(llrbIntItem(key))
				}
			}
		})
	}
}

func BenchmarkGetHit(b *testing.B) {
	for _, size := range querySizes {
		keys := shuffledRange(size, 2026)
		queries := sampleFrom(keys, min(size, 50_000), 7)

		b.Run(fmt.Sprintf("BTree/N=%d", size), func(b *testing.B) {
			tree := newBTree(keys)
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				if tree.Get(btreeIntItem(queries[i%len(queries)])) == nil {
					b.Fatal("missing key from btree")
				}
			}
		})

		b.Run(fmt.Sprintf("LLRB/N=%d", size), func(b *testing.B) {
			tree := newLLRB(keys)
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				if tree.Get(llrbIntItem(queries[i%len(queries)])) == nil {
					b.Fatal("missing key from llrb")
				}
			}
		})
	}
}

func BenchmarkGetMiss(b *testing.B) {
	for _, size := range querySizes {
		keys := shuffledRange(size, 2026)
		queries := missRange(size, min(size, 50_000))

		b.Run(fmt.Sprintf("BTree/N=%d", size), func(b *testing.B) {
			tree := newBTree(keys)
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				if tree.Get(btreeIntItem(queries[i%len(queries)])) != nil {
					b.Fatal("unexpected key in btree")
				}
			}
		})

		b.Run(fmt.Sprintf("LLRB/N=%d", size), func(b *testing.B) {
			tree := newLLRB(keys)
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				if tree.Get(llrbIntItem(queries[i%len(queries)])) != nil {
					b.Fatal("unexpected key in llrb")
				}
			}
		})
	}
}

func BenchmarkDeleteRandom(b *testing.B) {
	for _, size := range deleteSizes {
		keys := shuffledRange(size, 2026)

		b.Run(fmt.Sprintf("BTree/N=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				tree := newBTree(keys)
				for _, key := range keys {
					tree.Delete(btreeIntItem(key))
				}
			}
		})

		b.Run(fmt.Sprintf("LLRB/N=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				tree := newLLRB(keys)
				for _, key := range keys {
					tree.Delete(llrbIntItem(key))
				}
			}
		})
	}
}

func newBTree(keys []int) *btree.BTree {
	tree := btree.New(btreeDegree)
	for _, key := range keys {
		tree.ReplaceOrInsert(btreeIntItem(key))
	}
	return tree
}

func newLLRB(keys []int) *llrb.LLRB {
	tree := llrb.New()
	for _, key := range keys {
		tree.ReplaceOrInsert(llrbIntItem(key))
	}
	return tree
}

func shuffledRange(n int, seed int64) []int {
	values := make([]int, n)
	for i := range values {
		values[i] = i
	}

	r := rand.New(rand.NewSource(seed))
	r.Shuffle(n, func(i, j int) {
		values[i], values[j] = values[j], values[i]
	})
	return values
}

func sampleFrom(values []int, n int, seed int64) []int {
	samples := make([]int, n)
	r := rand.New(rand.NewSource(seed))
	for i := 0; i < n; i++ {
		samples[i] = values[r.Intn(len(values))]
	}
	return samples
}

func missRange(start, n int) []int {
	values := make([]int, n)
	for i := 0; i < n; i++ {
		values[i] = start + i + 1
	}
	return values
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
