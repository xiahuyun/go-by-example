# etcd Txn Revision (Main/Sub) 示例

这个示例专门演示：

1. 在 **一个 etcd 事务**里执行多个请求（`put/put/put/del`）
2. 通过 API 看到的 revision（只有 `main`）
3. 通过底层 backend（`member/snap/db`）解码出 `rev(main, sub)`

## 背景

etcd 内部 revision 结构是 `(main, sub)`：

- `main`：一次原子事务对应一个 main revision
- `sub`：同一个事务内每个写操作的顺序号（0,1,2...）

但客户端 API（`Get` / `Txn` 返回的 `mod_revision`）只暴露 `main`，不会直接返回 `sub`。

## 目录

- [main.go](/Users/hxia/.codex/worktrees/5d20/go-by-example/etcd/txn-rev-main-sub/main.go)
- [run-with-local-etcd.sh](/Users/hxia/.codex/worktrees/5d20/go-by-example/etcd/txn-rev-main-sub/run-with-local-etcd.sh)

## 运行

```bash
cd /Users/hxia/.codex/worktrees/5d20/go-by-example/etcd/txn-rev-main-sub
go mod tidy
./run-with-local-etcd.sh
```

或：

```bash
cd /Users/hxia/.codex/worktrees/5d20/go-by-example/etcd/txn-rev-main-sub
make demo
```

## 你会看到的输出重点

程序会打印两部分：

1. API 视角：
   - `response header revision(main)`
   - 每个 key 的 `createRev/modRev/version`

2. Backend 视角：
   - `rev(main=..., sub=..., tombstone=...)`
   - 对应 key/value

你会观察到：

- 同一事务里的多次写，`main` 相同
- `sub` 随写操作递增
- API 中 `mod_revision` 只有 `main`，看不到 `sub`

## 可选环境变量

- `ETCD_ENDPOINTS`：默认 `127.0.0.1:12379`
- `ETCD_DATA_DIR`：默认 `.tmp/etcd-data`（脚本里临时 etcd 的数据目录）
- `ETCD_SNAPSHOT`：默认 `.tmp/txn-rev-snapshot.db`（程序解析 main/sub 用的快照文件）
- `DEMO_PREFIX`：默认 `demo/txn-rev`
- `KEEP_ETCD_DATA=1`：脚本结束后保留数据目录
