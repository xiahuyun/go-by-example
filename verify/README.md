# verify：`go.etcd.io/etcd/client/pkg/v3/verify` 断言示例

这个示例演示 etcd 客户端工具包中的 `verify.Assert` 用法：  
在关键前置条件不满足时，立即触发断言并中止程序。

当前代码里会故意传入一个 `nil` 的 `Logger`，从而触发 panic，帮助理解断言失败的行为。

## 快速运行

```bash
cd /Users/hxia/project/go-by-example/verify
go run .
```

预期输出（节选）：

```text
panic: assertion failed: Logger should not be nil
```

## 代码说明

- `main` 中声明 `var lg *Logger`，默认是 `nil`
- `handle(lg)` 内调用：`verify.Assert(lg != nil, "Logger should not be nil")`
- 因条件为 `false`，程序触发断言失败并 panic

## 可以继续扩展

把 `lg` 改成非空实例，例如 `lg := &Logger{}`，再运行可观察断言通过时的行为。
