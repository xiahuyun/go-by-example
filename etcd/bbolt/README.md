# bbolt CRUD 使用示例（`go.etcd.io/bbolt`）

这个示例展示如何使用 `go.etcd.io/bbolt` 完成一个嵌入式 KV 数据库的常见操作：

- 建桶（`CreateBucketIfNotExists`）
- 增/改（`Put`）
- 查（`Get`）
- 删（`Delete`）
- 遍历（`Cursor`）

示例将 `User` 结构体以 JSON 形式存入 `users` bucket。

## 目录结构

- `main.go`：完整可运行示例
- `go.mod`：模块与依赖定义

## 运行方式

```bash
cd /Users/hxia/project/go-by-example/etcd/bbolt
go mod tidy
go run .
```

运行后会在当前目录生成一个本地数据库文件 `demo.db`。

## 代码要点

1. `bolt.Open` 打开/创建本地数据库文件。
2. 读写必须在事务中执行：
   - 写事务：`db.Update(...)`
   - 读事务：`db.View(...)`
3. `CreateBucketIfNotExists` 保证 bucket 初始化幂等。
4. `Cursor` 用于顺序遍历 bucket 中所有 key/value。

## 预期输出（示例）

```text
== Create: 插入数据 ==
Key=1001 User={ID:1001 Name:Alice Score:88}
Key=1002 User={ID:1002 Name:Bob Score:76}
Key=1003 User={ID:1003 Name:Cindy Score:93}

== Read: 按 ID 查询 ==
查询成功: {ID:1002 Name:Bob Score:76}

== Update: 更新 ID=1002 的分数 ==
更新后: {ID:1002 Name:Bob Score:84}

== Delete: 删除 ID=1001 ==
删除成功: ID=1001

== Read: 删除后再次查询 ID=1001 ==
查询结果: ID=1001 不存在

== 当前全部数据 ==
Key=1002 User={ID:1002 Name:Bob Score:84}
Key=1003 User={ID:1003 Name:Cindy Score:93}
```
