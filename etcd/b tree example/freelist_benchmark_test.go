package main

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/google/btree"
)

type intItem int

func (a intItem) Less(b btree.Item) bool {
	return a < b.(intItem)
}

const (
	benchDegree        = 32
	benchSeed    int64 = 20260408
	benchDataSize      = 10_000
	benchQuerySize     = 50_000
	benchChurnSize     = 10_000
	withFreeListSize   = 2_048
)

func BenchmarkFreeListInsertDeleteCycle(b *testing.B) {
	keys := shuffledKeys(benchDataSize, benchSeed)

	cases := []struct {
		name    string
		newTree func() *btree.BTree
	}{
		{
			name: "WithFreeList",
			newTree: func() *btree.BTree {
				return btree.NewWithFreeList(benchDegree, btree.NewFreeList(withFreeListSize))
			},
		},
		{
			name: "WithoutFreeList",
			newTree: func() *btree.BTree {
				// capacity=0 means freed nodes are never stored for reuse.
				return btree.NewWithFreeList(benchDegree, btree.NewFreeList(0))
			},
		},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			tree := tc.newTree()
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				for _, key := range keys {
					tree.ReplaceOrInsert(intItem(key))
				}
				for _, key := range keys {
					tree.Delete(intItem(key))
				}
			}

			b.StopTimer()
			if tree.Len() != 0 {
				b.Fatalf("tree should be empty after each insert/delete cycle, got len=%d", tree.Len())
			}
		})
	}
}

func BenchmarkFreeListGetHit(b *testing.B) {
	keys := shuffledKeys(benchDataSize, benchSeed)
	queries := sampleFromKeys(keys, benchQuerySize, benchSeed+7)

	cases := []struct {
		name    string
		newTree func() *btree.BTree
	}{
		{
			name: "WithFreeList",
			newTree: func() *btree.BTree {
				return btree.NewWithFreeList(benchDegree, btree.NewFreeList(withFreeListSize))
			},
		},
		{
			name: "WithoutFreeList",
			newTree: func() *btree.BTree {
				return btree.NewWithFreeList(benchDegree, btree.NewFreeList(0))
			},
		},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			tree := tc.newTree()
			for _, key := range keys {
				tree.ReplaceOrInsert(intItem(key))
			}

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				if tree.Get(intItem(queries[i%len(queries)])) == nil {
					b.Fatal("missing key from btree")
				}
			}
		})
	}
}

func BenchmarkFreeListDeleteThenWriteChurn(b *testing.B) {
	cases := []struct {
		name    string
		newTree func() *btree.BTree
	}{
		{
			name: "WithFreeList",
			newTree: func() *btree.BTree {
				return btree.NewWithFreeList(benchDegree, btree.NewFreeList(withFreeListSize))
			},
		},
		{
			name: "WithoutFreeList",
			newTree: func() *btree.BTree {
				return btree.NewWithFreeList(benchDegree, btree.NewFreeList(0))
			},
		},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			tree := tc.newTree()
			active := make([]int, benchChurnSize)
			for i := 0; i < benchChurnSize; i++ {
				active[i] = i
				tree.ReplaceOrInsert(intItem(i))
			}
			nextKey := benchChurnSize

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				idx := i % benchChurnSize
				oldKey := active[idx]
				if tree.Delete(intItem(oldKey)) == nil {
					b.Fatalf("delete miss for key=%d", oldKey)
				}

				newKey := nextKey
				nextKey++
				tree.ReplaceOrInsert(intItem(newKey))
				active[idx] = newKey
			}

			b.StopTimer()
			if tree.Len() != benchChurnSize {
				b.Fatalf("tree len mismatch after churn: got=%d want=%d", tree.Len(), benchChurnSize)
			}
		})
	}
}

func shuffledKeys(n int, seed int64) []int {
	values := make([]int, n)
	for i := range values {
		values[i] = i
	}

	r := rand.New(rand.NewSource(seed))
	r.Shuffle(len(values), func(i, j int) {
		values[i], values[j] = values[j], values[i]
	})
	return values
}

func sampleFromKeys(values []int, n int, seed int64) []int {
	if len(values) == 0 {
		panic("sample source must not be empty")
	}
	out := make([]int, n)
	r := rand.New(rand.NewSource(seed))
	for i := 0; i < n; i++ {
		out[i] = values[r.Intn(len(values))]
	}
	return out
}

func BenchmarkFreeListStats(b *testing.B) {
	// A compact summary benchmark to quickly compare write-heavy behavior.
	for _, n := range []int{1_000, 10_000} {
		b.Run(fmt.Sprintf("InsertDelete/N=%d", n), func(b *testing.B) {
			keys := shuffledKeys(n, benchSeed+int64(n))
			run := func(tree *btree.BTree) {
				for _, key := range keys {
					tree.ReplaceOrInsert(intItem(key))
				}
				for _, key := range keys {
					tree.Delete(intItem(key))
				}
			}

			b.Run("WithFreeList", func(b *testing.B) {
				tree := btree.NewWithFreeList(benchDegree, btree.NewFreeList(withFreeListSize))
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					run(tree)
				}
			})

			b.Run("WithoutFreeList", func(b *testing.B) {
				tree := btree.NewWithFreeList(benchDegree, btree.NewFreeList(0))
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					run(tree)
				}
			})
		})
	}
}
