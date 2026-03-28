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
