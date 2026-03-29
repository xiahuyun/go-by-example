# raft-wal-restart：WAL 持久化 + 重启恢复示例

这个示例基于 `go.etcd.io/raft/v3`，在单进程里模拟 3 节点 Raft 集群，并演示：

- 首次启动：`StartNode`
- 将 `HardState + Entries` 追加到 WAL
- 定期生成状态机快照并保存
- 全部节点停止后，用磁盘数据恢复：`RestartNode`
- 恢复后继续写入，验证集群可继续工作

## 运行

```bash
cd /Users/hxia/project/go-by-example/etcd/raft-wal-restart
go mod tidy
go run .
```

## 你会看到什么

程序分两阶段输出：

1. `phase 1`：新集群选主并写入 `x/y/user`
2. `phase 2`：重启后自动走 `RestartNode`，验证旧数据恢复，再写入新值 `z`
3. 再停掉一个 follower，验证多数派仍可提交 `k=v`

## 关键 API 对应关系

- 启动新节点：`raft.StartNode`
- 重启恢复：`raft.RestartNode`
- 时钟推进：`node.Tick`
- 主循环：`node.Ready` -> 持久化 -> `Step` 转发 -> 应用 `CommittedEntries` -> `Advance`
- 客户端写入：`node.Propose`

## 快照后的压缩（compactTo）

示例里在生成快照后会做一次日志压缩，逻辑是：

```go
compactTo := uint64(1)
if snap.Metadata.Index > n.snapshotCatchUp {
	compactTo = snap.Metadata.Index - n.snapshotCatchUp
}
n.storage.Compact(compactTo)
```

这段代码的目的：

- 回收旧日志，降低内存/WAL 体积
- 不直接压到快照点，而是保留 `snapshotCatchUp` 条最近日志，给慢 follower 一个“日志追赶窗口”
- 避免 follower 一落后就必须发快照，减少快照传输频率

举例：

- 如果 `snap.Metadata.Index=10`，`snapshotCatchUp=2`
- 那么 `compactTo=8`
- 表示把更老日志压缩掉，保留最近一小段日志用于追赶

错误处理里忽略 `ErrCompacted`（说明已经压过）属于正常的幂等行为。

## 磁盘文件

运行后会在当前目录生成：

- `./data/node-1/wal.log`
- `./data/node-1/snapshot.bin`
- `./data/node-2/...`
- `./data/node-3/...`

说明：

- `wal.log` 是简化版 WAL（按行 JSON，保存 `HardState` 和 `Entries`）
- `snapshot.bin` 是 `raftpb.Snapshot` 二进制

这是教学实现，重点是帮助你理解 Raft 接入流程，不是生产级 WAL 实现。
