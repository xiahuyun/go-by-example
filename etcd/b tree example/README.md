# B-Tree CRUD 使用示例（`github.com/google/btree`）

本示例展示如何使用 `github.com/google/btree` 完成常见的增删改查（CRUD）操作：

- 增（Create）：`ReplaceOrInsert`
- 查（Read）：`Get`
- 改（Update）：`ReplaceOrInsert`（同 key 覆盖）
- 删（Delete）：`Delete`

示例数据是一个按 `ID` 排序的用户集合（`User`）。

## 目录结构

- `main.go`：完整可运行示例
- `go.mod`：模块与依赖定义
- `freelist_benchmark_test.go`：`freelist` 与 `no-freelist` 的性能对比基准

## 运行方式

```bash
cd '/Users/hxia/project/go-by-example/etcd/b tree example'
go mod tidy
go run .
```

## 代码要点

1. `UserItem` 实现 `btree.Item` 接口中的 `Less` 方法，用于定义排序规则（按 `ID` 升序）。
2. `upsertUser` 使用 `ReplaceOrInsert`：
   - key 不存在：插入
   - key 已存在：覆盖旧值（用于更新）
3. `getUser` 使用 `Get` 按 key 精确查询。
4. `deleteUser` 使用 `Delete` 删除指定 key。
5. `printAll` 使用 `Ascend` 按序遍历。

## 预期输出（示例）

```text
== Create: 插入数据 ==
ID=1001 Name=Alice Score=88
ID=1002 Name=Bob Score=76
ID=1003 Name=Cindy Score=93

== Read: 按 ID 查询 ==
查询成功: {ID:1002 Name:Bob Score:76}

== Read: 按 用户名 查询 ==
查询失败: ID=1002 不存在

== Update: 更新 ID=1002 的分数 ==
更新后: {ID:1002 Name:Bob Score:84}

== Delete: 删除 ID=1001 ==
删除成功: ID=1001

== Read: 删除后再次查询 ID=1001 ==
查询结果: ID=1001 不存在

== 当前全部数据（有序） ==
ID=1002 Name=Bob Score=84
ID=1003 Name=Cindy Score=93
```

## freelist 性能对比（Benchmark）

运行：

```bash
cd '/Users/hxia/project/go-by-example/etcd/b tree example'
go test -run=^$ -bench=FreeList -benchmem
```

说明：

- `WithFreeList`：使用 `btree.NewWithFreeList(..., btree.NewFreeList(2048))`
- `WithoutFreeList`：使用 `btree.NewWithFreeList(..., btree.NewFreeList(0))`（等价于禁用节点复用）
- `InsertDeleteCycle` / `Stats`：写密集场景（最容易体现 freelist 价值）
- `DeleteThenWriteChurn`：持续“删旧写新”的抖动写入场景（维持固定工作集大小）
- `GetHit`：读密集场景（用于对照）

### 一次实测结果

测试环境：

- `goos=darwin`
- `goarch=arm64`
- `cpu=Apple M3`
- 命令：`go test -run=^$ -bench=FreeList -benchmem`

结果（节选）：

| Benchmark | WithFreeList | WithoutFreeList |
| --- | ---: | ---: |
| `BenchmarkFreeListInsertDeleteCycle` | `4,663,481 ns/op` / `158,153 B/op` / `19,493 allocs/op` | `10,302,232 ns/op` / `537,817 B/op` / `20,175 allocs/op` |
| `BenchmarkFreeListGetHit` | `326.2 ns/op` / `7 B/op` / `0 allocs/op` | `571.8 ns/op` / `7 B/op` / `0 allocs/op` |
| `BenchmarkFreeListDeleteThenWriteChurn` | `421.9 ns/op` / `16 B/op` / `2 allocs/op` | `429.9 ns/op` / `72 B/op` / `2 allocs/op` |
| `BenchmarkFreeListStats/InsertDelete/N=1000` | `412,984 ns/op` / `11,918 B/op` / `1,488 allocs/op` | `635,927 ns/op` / `52,064 B/op` / `1,568 allocs/op` |
| `BenchmarkFreeListStats/InsertDelete/N=10000` | `6,422,379 ns/op` / `158,082 B/op` / `19,493 allocs/op` | `7,222,183 ns/op` / `539,537 B/op` / `20,178 allocs/op` |

结果说明：

- 写密集场景（`InsertDeleteCycle` / `Stats`）下，`WithFreeList` 明显更快、且 `B/op` 显著更低，说明节点复用有效降低了重复分配和 GC 压力。
- `DeleteThenWriteChurn` 这种“删后立刻写”的持续抖动场景中，`WithFreeList` 的 `B/op` 也更低（`16` vs `72`），说明即使延迟差距不大，内存分配压力仍然明显下降。
- 在 `N=10000` 时，`WithFreeList` 的内存占用约为 `WithoutFreeList` 的 `~29%`（`158KB` vs `540KB`），这是 freelist 价值最直观的体现。
- 读密集场景（`GetHit`）理论上不直接依赖 freelist；该项主要作为对照。若要得到更稳定结论，建议多次运行并取中位数（例如 `-count=5`）。
