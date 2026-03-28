# Async Start Example (etcd style)

这个示例模仿了 etcd 的异步启动结构：

- `StartAsyncServer` 类似 `embed.StartEtcd`
- `startAsyncServer` 里注册了中断关闭钩子，随后等待 `ReadyNotify` / `StopNotify`
- 返回 `StopNotify` 和 `Err` 两个通道供调用方监听

## 运行

```bash
cd /Users/hxia/project/go-by-example/etcd/async-start
go run .
```

## 可调整项

在 `main.go` 的 `ServerConfig` 中可调整：

- `StartupDelay`：启动耗时
- `TickInterval`：运行期任务节奏
- `RunDuration`：自动优雅退出时间
- `FailStartup`：是否模拟启动失败
