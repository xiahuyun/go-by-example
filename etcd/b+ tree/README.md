# etcd Snapshot B+Tree Visualizer (Go)

这个工具会执行以下流程：

1. 调用 `etcdutl snapshot dump` 读取 snapshot
2. 解析 `rev/key/value`
3. 按 `revision` 排序
4. 按页大小模拟 B+ 树叶子节点
5. 生成 Graphviz `.dot`

## 文件

- `etcd_bptree_viz.go`

## 使用

在当前目录（`/Users/hxia/project/go-by-example/etcd`）执行：

```bash
cd 'b+ tree'
go build -o viz etcd_bptree_viz.go
```

准备 snapshot（示例）：

```bash
etcdctl snapshot save snapshot.db
```

运行可视化：

```bash
./viz -snapshot snapshot.db -page-size 4 -out tree.dot
```

可选参数：

- `-bucket`：默认 `key`（etcd mvcc bucket）
- `-dump-out`：导出中间 dump 文本（调试用）
- `-show-value`：在叶子节点显示 value

渲染 PNG：

```bash
dot -Tpng tree.dot -o tree.png
```
