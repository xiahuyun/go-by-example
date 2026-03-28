# BTree 与红黑树（LLRB）性能对比报告

## 测试环境

- 日期：2026-03-28
- 机器：Apple M3
- OS：darwin / arm64
- Go 包：`github.com/google/btree` vs `github.com/petar/GoLLRB/llrb`
- 命令：`GOCACHE=/tmp/go-build go test -run=^$ -bench=. -benchmem -count=3`
- 统计方式：取 3 次结果的中位数（`benchmark_median.txt`）

## 结果总览（中位数）

| 场景 | N | BTree ns/op | LLRB ns/op | 更快方 | 速度差距 |
|---|---:|---:|---:|---|---:|
| InsertRandom | 10,000 | 4,034,197 | 5,655,777 | BTree | 1.40x |
| InsertRandom | 100,000 | 80,652,772 | 107,854,564 | BTree | 1.34x |
| GetHit | 10,000 | 374.90 | 369.70 | LLRB | 1.01x |
| GetHit | 100,000 | 744.00 | 1,147.00 | BTree | 1.54x |
| GetMiss | 10,000 | 167.40 | 136.50 | LLRB | 1.23x |
| GetMiss | 100,000 | 215.00 | 181.20 | LLRB | 1.19x |
| DeleteRandom | 10,000 | 8,596,156 | 12,735,841 | BTree | 1.48x |
| DeleteRandom | 50,000 | 60,825,273 | 83,054,854 | BTree | 1.37x |

## 内存分配对比（中位数）

| 场景 | N | BTree B/op | LLRB B/op | BTree allocs/op | LLRB allocs/op |
|---|---:|---:|---:|---:|---:|
| InsertRandom | 10,000 | 473,924 | 557,967 | 10,468 | 19,744 |
| InsertRandom | 100,000 | 4,704,999 | 5,597,992 | 106,705 | 199,744 |
| DeleteRandom | 10,000 | 556,486 | 635,915 | 20,216 | 29,488 |
| DeleteRandom | 50,000 | 2,767,650 | 3,195,957 | 102,974 | 149,488 |
| GetHit / GetMiss | 10,000 / 100,000 | 基本一致 | 基本一致 | 基本一致 | 基本一致 |

## 结论

- 写入和删除场景：BTree 明显更快，且分配更少。
- 点查场景：`GetHit(100k)` BTree 更快，`GetHit(10k)` 两者基本持平；`GetMiss` 场景 LLRB 更快。
- 结论上可视为“BTree 写入/删除优势明显，读路径互有胜负”。
- 如果你的场景是“写多 + 内存敏感”，优先选 BTree。
- 如果是“读多 + 数据规模较小 + 实现简单优先”，LLRB 也有竞争力。

## 复现方式

```bash
cd '/Users/hxia/project/go-by-example/etcd/b tree'
./run_bench.sh
```

运行后查看：

- 原始数据：`benchmark_count3.txt`
- 中位数汇总：`benchmark_median.txt`
- 说明：benchmark 对系统负载敏感，建议在机器空闲时重复执行 2-3 次并观察趋势。
