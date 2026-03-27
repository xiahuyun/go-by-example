# gRPC Round Robin Client Example

这个示例包含：

- 两个 gRPC server 节点（`127.0.0.1:50051`、`127.0.0.1:50052`）
- 一个客户端，使用 `round_robin` 负载策略轮询访问后端

## 运行方式

在 `/Users/hxia/project/go-by-example/etcd/grpc` 目录执行：

```bash
go mod tidy
go run ./server
```

另开一个终端执行：

```bash
go run ./client
```

如果你看到客户端输出在 `node-1` 和 `node-2` 之间交替，就说明 `round_robin` 生效。
