# BTree vs 红黑树（LLRB）基准测试

本目录用于对比两种 Go 树结构实现的性能：

- BTree: <http://godoc.org/github.com/google/btree>
- 红黑树（LLRB）: <http://godoc.org/github.com/petar/GoLLRB/llrb>

## 文件说明

- `tree_benchmark_test.go`: 基准测试代码（插入 / 查询命中 / 查询未命中 / 删除）
- `run_bench.sh`: 一键跑 3 轮 benchmark，并生成中位数汇总
- `benchmark_count3.txt`: 3 轮原始结果（运行后生成）
- `benchmark_median.txt`: 中位数结果（运行后生成）
- `BENCHMARK_REPORT.md`: 本机一次实测报告与结论

## 如何运行

```bash
cd '/Users/hxia/project/go-by-example/etcd/b tree'
go mod tidy
chmod +x run_bench.sh
./run_bench.sh
```

如果你只想快速跑一轮：

```bash
GOCACHE=/tmp/go-build go test -run=^$ -bench=. -benchmem
```

## 基准测试设计

- 数据集：`[0, N)` 的唯一整数，固定随机种子打乱，确保可复现
- 规模（`InsertRandom`）：N=10,000 / 100,000
- 规模（`GetHit`）：N=10,000 / 100,000
- 规模（`GetMiss`）：N=10,000 / 100,000
- 规模（`DeleteRandom`）：N=10,000 / 50,000
- BTree 配置：`degree=32`
- 输出指标：`ns/op`、`B/op`、`allocs/op`

## 如何直观看对比

优先看 `BENCHMARK_REPORT.md` 的汇总表（已给出中位数和胜出方），再结合 `benchmark_count3.txt` 看波动范围。
