# etcd MVCC 示例

这个示例演示 etcd 的 MVCC（多版本并发控制）核心能力：

- 同一个 key 的多版本写入
- 按 revision 读取历史快照（`WithRev`）
- 基于 `ModRevision` 的 CAS（乐观并发控制）
- `compact` 后旧 revision 不可读（`ErrCompacted`）

## 目录

- [main.go](/Users/hxia/project/go-by-example/etcd/mvcc/main.go)
- [run-with-local-etcd.sh](/Users/hxia/project/go-by-example/etcd/mvcc/run-with-local-etcd.sh)

## 运行前准备

确保本地有可访问的 etcd（默认会连）：

- `localhost:2379`
- `localhost:22379`
- `localhost:32379`

如果你只有单节点 etcd，可以通过环境变量覆盖：

```bash
export ETCD_ENDPOINTS=localhost:2379
```

另外，使用一键脚本时需要本机可执行 `etcd` 命令（或通过 `ETCD_BIN` 指定路径）。

## 运行

```bash
cd /Users/hxia/project/go-by-example/etcd/mvcc
go mod tidy
go run .
```

或者用 Makefile：

```bash
cd /Users/hxia/project/go-by-example/etcd/mvcc
make run
```

## 一键启动本地 etcd 并运行示例

```bash
cd /Users/hxia/project/go-by-example/etcd/mvcc
./run-with-local-etcd.sh
```

或者更短：

```bash
cd /Users/hxia/project/go-by-example/etcd/mvcc
make mvcc-demo
```

脚本行为：

- 若 `127.0.0.1:12379` 已有健康 etcd，则直接复用
- 否则拉起一个临时单节点 etcd（数据目录：`.tmp/etcd-data`）
- 等待健康后自动执行 `go run .`
- 结束后自动停止脚本拉起的 etcd

常用可选环境变量：

```bash
# 指定 etcd 可执行文件
ETCD_BIN=/opt/homebrew/bin/etcd ./run-with-local-etcd.sh

# 指定端口（默认 12379/12380）
ETCD_PORT=12379 ETCD_PEER_PORT=12380 ./run-with-local-etcd.sh

# 保留数据目录（默认 0，表示清理）
KEEP_ETCD_DATA=1 ./run-with-local-etcd.sh
```

`make mvcc-demo` 同样支持这些环境变量，例如：

```bash
cd /Users/hxia/project/go-by-example/etcd/mvcc
ETCD_PORT=12379 ETCD_PEER_PORT=12380 KEEP_ETCD_DATA=1 make mvcc-demo
```

## 你会看到什么

程序会按步骤打印：

1. 连续 3 次 `Put`，每次 cluster revision 递增
2. 在 `rev1/rev2/rev3` 上分别 `Get`，拿到不同版本值
3. 用最新 `ModRevision` 做一次 CAS（成功）
4. 用旧 revision 做一次 CAS（失败）
5. 从指定 revision 开始 `Watch`，观察事件回放
6. `Compact` 后读取旧 revision，得到 `ErrCompacted`

这就是 etcd MVCC 的关键行为闭环。
