# raft：`go.etcd.io/raft/v3` 最小可运行示例

这个示例用**单进程模拟 3 节点 Raft 集群**，目标是让你看到 `go.etcd.io/raft/v3` 的核心用法和调用顺序。

它演示了：

- 3 节点启动并选主
- leader 提案（`Propose`）写入一个简单 KV 状态机
- 日志复制与提交（所有存活节点应用到同样的值）
- 停掉一个 follower 后，剩余多数派（2/3）仍可提交新日志

## 快速运行

```bash
cd /Users/hxia/project/go-by-example/etcd/raft
go mod tidy
go run .
```

你会看到类似输出（节选）：

```text
leader elected: node 2
[node 2] apply index=3 term=2 set x=10
[node 1] apply index=3 term=2 set x=10
[node 3] apply index=3 term=2 set x=10
...
stop one follower node 1 (simulate node down)
with one node down, remaining majority still commits:
node=2 appliedIndex=5 kv={x=10, y=20, z=30}
node=3 appliedIndex=5 kv={x=10, y=20, z=30}
```

## 代码里对应的 Raft 用法

`main.go` 里最关键的是下面这条链路：

1. 初始化节点：`raft.StartNode(cfg, peers)`
2. 周期推进时钟：`node.Tick()`
3. 消费 `Ready()` 批次：
   - 持久化 `HardState/Entries`（示例里用 `MemoryStorage`）
   - 发送网络消息（示例里是内存转发到 `Step`）
   - 应用 `CommittedEntries` 到状态机
   - `node.Advance()` 通知 Raft 此批处理完成
4. 客户端写入：对 leader 调用 `node.Propose(...)`

这就是使用 `go.etcd.io/raft/v3` 的核心循环。

## 生产环境要补齐什么

这个示例是教学版，省略了不少工程细节。生产环境你通常还需要：

- 持久化 WAL 与快照（不是 `MemoryStorage`）
- 真正的网络传输层（RPC）
- 节点重启恢复、快照安装、日志压缩
- 成员变更（`ConfChangeV2`）的完整流程
