# etcd-bbolt：直接读取 etcd 底层数据的小示例

这个项目演示了：**如何使用 `go.etcd.io/bbolt` 直接打开 etcd 的底层 `db` 文件，并解析出真实的业务 key/value**。

换句话说，它不是通过 etcd 的 HTTP/gRPC API 读数据，而是直接从 etcd 的存储文件（`member/snap/db`）里按 revision 查询，再把 protobuf 数据反序列化成可读内容。

## 它是做什么的

- 用于理解 etcd 数据在 bbolt 中的存储形式
- 用于离线排查某个 revision 对应的数据内容
- 用于学习 etcd `mvccpb.KeyValue` 的二进制结构与解码方式

## 当前示例做了什么

`main.go` 中主要流程如下：

1. 打开 etcd 的 bbolt 文件：`.../member/snap/db`
2. 将一个十六进制 revision key 转为二进制
3. 在 bucket `key` 中按该 revision key 查找 value
4. 将 value 反序列化为 `mvccpb.KeyValue`
5. 打印用户 key、value 和 `ModRevision`

## 适合谁

- 正在学习 etcd 存储层（MVCC + bbolt）的开发者
- 需要做底层数据诊断或问题定位的同学

## 运行方式

```bash
cd /Users/hxia/project/go-by-example/etcd/etcd-bbolt
go mod tidy
go run .
```

## 注意事项

- 示例中的 etcd 数据文件路径是本地绝对路径，请按你的环境修改。
- 示例中的 `keyHex` 需要替换成你要查询的 revision key。
- 直接读取线上 etcd 数据文件要谨慎，建议先在测试/备份环境验证。

