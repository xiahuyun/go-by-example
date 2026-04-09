# COW（Copy-On-Write）示例

`COW` 是 `Copy-On-Write`，中文常说“写时复制”：

- 读时共享同一份底层数据（不复制）
- 写时只复制被修改路径上的节点（而不是整棵树）

`github.com/google/btree` 的 `Clone()` 就是这个思路：  
克隆后两棵树先共享旧节点，后续各自写入时按路径复制（通过 `copyOnWriteContext` + `mutableFor` 判断所有权）。

## 本示例演示点

1. `Clone` 后 `origin` 与 `clone` 共享同一棵旧树。
2. 写 `clone` 的左子树时，只复制 `root + left` 路径，`right` 仍共享。
3. 再写 `origin` 的右子树时，继续只复制各自写路径。

## 运行

```bash
cd /Users/hxia/project/go-by-example/cow
go run .
```
